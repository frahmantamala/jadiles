package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	"github.com/frahmantamala/jadiles/internal/services"
	"github.com/frahmantamala/jadiles/internal/services/booking"
	"gorm.io/gorm"
)

// CreateBookingWithTransaction creates a booking with sessions atomically
// Uses pessimistic locking (SELECT FOR UPDATE) to prevent double bookings
func (r *Repository) CreateBookingWithTransaction(ctx context.Context, req *services.CreateBookingRequest) (*services.Booking, error) {
	var booking *services.Booking

	// Execute everything in a transaction
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Verify child belongs to parent
		var child datamodel.Children
		if err := tx.Where("id = ? AND parent_id = ?", req.ChildID, req.ParentID).First(&child).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return internal.NewNotFoundError("Child not found or does not belong to parent")
			}
			return internal.NewInternalServerError(err)
		}

		// 2. Get service with vendor info
		var service datamodel.Services
		if err := tx.Where("id = ? AND status = ?", req.ServiceID, "active").First(&service).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return internal.NewNotFoundError("Service not found or inactive")
			}
			return internal.NewInternalServerError(err)
		}

		// 3. Check and reserve slots with pessimistic locking
		if err := checkAndReserveSlots(tx, req.SessionDates); err != nil {
			return err
		}

		// 4. Calculate total amount
		totalAmount, err := calculateBookingAmount(&service, req.BookingType)
		if err != nil {
			return err
		}

		// 5. Generate unique booking number
		bookingNumber := generateBookingNumber()

		// 6. Create booking record
		bookingData := &datamodel.Booking{
			BookingNumber:  bookingNumber,
			ParentID:       req.ParentID,
			ChildID:        req.ChildID,
			ServiceID:      req.ServiceID,
			VendorID:       service.VendorID,
			BookingType:    string(req.BookingType),
			TotalSessions:  req.BookingType.GetSessionCount(),
			TotalAmount:    totalAmount,
			Status:         string(services.BookingStatusPending),
			PreferredCoach: req.PreferredCoach,
			ParentNotes:    req.ParentNotes,
			Version:        1, // Initial version for optimistic locking
		}

		if err := tx.Create(bookingData).Error; err != nil {
			return err
		}

		// 7. Create booking sessions
		sessions, err := createBookingSessions(tx, bookingData.ID, req.SessionDates)
		if err != nil {
			return err
		}

		// 8. Build domain booking object
		booking = &services.Booking{
			ID:             bookingData.ID,
			BookingNumber:  bookingData.BookingNumber,
			ParentID:       bookingData.ParentID,
			ChildID:        bookingData.ChildID,
			ServiceID:      bookingData.ServiceID,
			VendorID:       bookingData.VendorID,
			BookingType:    services.BookingType(bookingData.BookingType),
			TotalSessions:  bookingData.TotalSessions,
			TotalAmount:    bookingData.TotalAmount,
			Status:         services.BookingStatus(bookingData.Status),
			PreferredCoach: bookingData.PreferredCoach,
			ParentNotes:    bookingData.ParentNotes,
			Version:        bookingData.Version,
			CreatedAt:      bookingData.CreatedAt,
			UpdatedAt:      bookingData.UpdatedAt,
			Sessions:       sessions,
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return booking, nil
}

// checkAndReserveSlots verifies slot availability with row-level locking
func checkAndReserveSlots(tx *gorm.DB, sessionDates []services.BookingSessionRequest) error {
	for i, session := range sessionDates {
		var schedule datamodel.Schedule

		// Use SELECT FOR UPDATE to lock the row and prevent concurrent bookings
		query := `
			SELECT id, service_id, day_of_week, start_time, end_time, available_slots, is_active
			FROM schedules
			WHERE id = ? AND is_active = true
			FOR UPDATE
		`

		if err := tx.Raw(query, session.ScheduleID).Scan(&schedule).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return internal.NewNotFoundError(fmt.Sprintf("Session %d: schedule not found or inactive", i+1))
			}
			return internal.NewInternalServerError(err)
		}

		// Count existing bookings for this schedule and date
		var bookedCount int64
		countQuery := `
			SELECT COUNT(*)
			FROM booking_sessions
			WHERE schedule_id = ?
			  AND session_date = ?
			  AND status NOT IN ('cancelled', 'no_show')
		`

		if err := tx.Raw(countQuery, session.ScheduleID, session.SessionDate).Scan(&bookedCount).Error; err != nil {
			return internal.NewInternalServerError(err)
		}

		// Check if slots are available
		availableSlots := schedule.AvailableSlots - int(bookedCount)
		if availableSlots <= 0 {
			return internal.NewConflictError(
				fmt.Sprintf("Session %d: no available slots for %s", i+1, session.SessionDate.Format("2006-01-02")),
				internal.ErrConflict,
			)
		}
	}

	return nil
}

// createBookingSessions creates all booking session records
func createBookingSessions(tx *gorm.DB, bookingID int64, sessionDates []services.BookingSessionRequest) ([]*services.BookingSession, error) {
	sessions := make([]*services.BookingSession, 0, len(sessionDates))

	for _, sessionReq := range sessionDates {
		// Get schedule details for times
		var schedule datamodel.Schedule
		if err := tx.Where("id = ?", sessionReq.ScheduleID).First(&schedule).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, internal.NewNotFoundError("Schedule not found")
			}
			return nil, internal.NewInternalServerError(err)
		}

		sessionData := &datamodel.BookingSession{
			BookingID:   bookingID,
			ScheduleID:  sessionReq.ScheduleID,
			SessionDate: sessionReq.SessionDate,
			Status:      string(services.SessionStatusScheduled),
			CoachID:     schedule.CoachID,
		}

		if err := tx.Create(sessionData).Error; err != nil {
			return nil, internal.NewInternalServerError(err)
		}

		sessions = append(sessions, &services.BookingSession{
			ID:          sessionData.ID,
			BookingID:   sessionData.BookingID,
			ScheduleID:  sessionData.ScheduleID,
			SessionDate: sessionData.SessionDate,
			StartTime:   schedule.StartTime,
			EndTime:     schedule.EndTime,
			Status:      services.SessionStatus(sessionData.Status),
			CoachID:     sessionData.CoachID,
			CreatedAt:   sessionData.CreatedAt,
			UpdatedAt:   sessionData.UpdatedAt,
		})
	}

	return sessions, nil
}

// calculateBookingAmount calculates total amount based on booking type
func calculateBookingAmount(service *datamodel.Services, bookingType services.BookingType) (float64, error) {
	switch bookingType {
	case services.BookingTypeTrial:
		if service.TrialPrice != nil {
			return *service.TrialPrice, nil
		}
		return 0, internal.NewValidationError("Trial price not available for this service")

	case services.BookingTypeSingle:
		return service.PricePerSession, nil

	case services.BookingTypePackage4:
		if service.Package4Price != nil {
			return *service.Package4Price, nil
		}
		return service.PricePerSession * 4, nil

	case services.BookingTypePackage8:
		if service.Package8Price != nil {
			return *service.Package8Price, nil
		}
		return service.PricePerSession * 8, nil

	case services.BookingTypePackage12:
		if service.Package12Price != nil {
			return *service.Package12Price, nil
		}
		return service.PricePerSession * 12, nil

	default:
		return 0, internal.NewValidationError(fmt.Sprintf("Invalid booking type: %s", bookingType))
	}
}

// generateBookingNumber generates a unique booking number
func generateBookingNumber() string {
	// Format: BK-YYYYMMDD-HHMMSS-RANDOM
	now := time.Now()
	return fmt.Sprintf("BK-%s-%06d", now.Format("20060102-150405"), now.Nanosecond()%1000000)
}

// GetServiceNameByID retrieves service name by ID
func (r *Repository) GetServiceNameByID(ctx context.Context, serviceID int64) (string, error) {
	var service datamodel.Services
	if err := r.db.WithContext(ctx).Select("name").Where("id = ?", serviceID).First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return service.Name, nil
}

// GetChildNameByID retrieves child name by ID
func (r *Repository) GetChildNameByID(ctx context.Context, childID int64) (string, error) {
	var child datamodel.Children
	if err := r.db.WithContext(ctx).Select("name").Where("id = ?", childID).First(&child).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return child.Name, nil
}

// GetCoachNameByID retrieves coach name by ID
func (r *Repository) GetCoachNameByID(ctx context.Context, coachID int64) (string, error) {
	var coach struct {
		FullName string `db:"full_name"`
	}
	if err := r.db.WithContext(ctx).Table("coaches").Select("full_name").Where("id = ?", coachID).Scan(&coach).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return coach.FullName, nil
}

// GetBookingByID retrieves a booking with all sessions
func (r *Repository) GetBookingByID(ctx context.Context, bookingID int64) (*services.Booking, error) {
	var bookingData datamodel.Booking
	if err := r.db.WithContext(ctx).Where("id = ?", bookingID).First(&bookingData).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	// Get sessions
	var sessionsData []datamodel.BookingSession
	if err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).Find(&sessionsData).Error; err != nil {
		return nil, err
	}

	// Convert to domain models
	sessions := make([]*services.BookingSession, 0, len(sessionsData))
	for _, sd := range sessionsData {
		// Get schedule for times
		var schedule datamodel.Schedule
		r.db.WithContext(ctx).Where("id = ?", sd.ScheduleID).First(&schedule)

		sessions = append(sessions, &services.BookingSession{
			ID:          sd.ID,
			BookingID:   sd.BookingID,
			ScheduleID:  sd.ScheduleID,
			SessionDate: sd.SessionDate,
			StartTime:   schedule.StartTime,
			EndTime:     schedule.EndTime,
			Status:      services.SessionStatus(sd.Status),
			CoachID:     sd.CoachID,
			CreatedAt:   sd.CreatedAt,
			UpdatedAt:   sd.UpdatedAt,
		})
	}

	booking := &services.Booking{
		ID:             bookingData.ID,
		BookingNumber:  bookingData.BookingNumber,
		ParentID:       bookingData.ParentID,
		ChildID:        bookingData.ChildID,
		ServiceID:      bookingData.ServiceID,
		VendorID:       bookingData.VendorID,
		BookingType:    services.BookingType(bookingData.BookingType),
		TotalSessions:  bookingData.TotalSessions,
		TotalAmount:    bookingData.TotalAmount,
		Status:         services.BookingStatus(bookingData.Status),
		PreferredCoach: bookingData.PreferredCoach,
		ParentNotes:    bookingData.ParentNotes,
		Version:        bookingData.Version,
		CreatedAt:      bookingData.CreatedAt,
		UpdatedAt:      bookingData.UpdatedAt,
		Sessions:       sessions,
	}

	return booking, nil
}

// GetBookingEnrichment fetches service, child, and vendor names for booking detail
func (r *Repository) GetBookingEnrichment(ctx context.Context, serviceID, childID, vendorID int64) (*booking.BookingEnrichment, error) {
	enrichment := &booking.BookingEnrichment{}

	// Fetch service name and category
	var serviceData struct {
		Name         string
		CategoryName string
	}
	err := r.db.WithContext(ctx).
		Table("services").
		Select("services.name, categories.name as category_name").
		Joins("LEFT JOIN categories ON categories.id = services.category_id").
		Where("services.id = ?", serviceID).
		Scan(&serviceData).Error

	if err == nil {
		enrichment.ServiceName = serviceData.Name
		enrichment.CategoryName = serviceData.CategoryName
	}

	// Fetch child name
	var child datamodel.Children
	if err := r.db.WithContext(ctx).Select("name").Where("id = ?", childID).First(&child).Error; err == nil {
		enrichment.ChildName = child.Name
	}

	// Fetch vendor name
	var vendor datamodel.Vendor
	if err := r.db.WithContext(ctx).Select("business_name").Where("id = ?", vendorID).First(&vendor).Error; err == nil {
		enrichment.VendorName = vendor.BusinessName
	}

	return enrichment, nil
}
