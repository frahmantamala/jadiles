package booking

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/services"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

type Repository interface {
	CreateBookingWithTransaction(ctx context.Context, req *services.CreateBookingRequest) (*services.Booking, error)
	GetBookingByID(ctx context.Context, bookingID int64) (*services.Booking, error)
	GetServiceNameByID(ctx context.Context, serviceID int64) (string, error)
	GetChildNameByID(ctx context.Context, childID int64) (string, error)
	GetCoachNameByID(ctx context.Context, coachID int64) (string, error)
	GetBookingEnrichment(ctx context.Context, serviceID, childID, vendorID int64) (*BookingEnrichment, error)
}

type ServiceUsecase struct {
	repo Repository
}

func NewService(repo Repository) *ServiceUsecase {
	return &ServiceUsecase{
		repo: repo,
	}
}

func (s *ServiceUsecase) CreateBooking(ctx context.Context, req *services.CreateBookingRequest) (*services.BookingConfirmation, error) {
	if err := req.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	now := time.Now()
	for i, session := range req.SessionDates {
		if session.SessionDate.Before(now) {
			return nil, internal.NewValidationError(fmt.Sprintf("session %d: cannot book sessions in the past", i+1))
		}
	}

	const maxRetries = 3
	var booking *services.Booking
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		booking, err = s.repo.CreateBookingWithTransaction(ctx, req)
		if err == nil {
			break // Success
		}

		var appErr *internal.AppError
		if internal.IsAppError(err) {
			appErr = internal.GetAppError(err)
			// Retry on conflict errors
			if appErr.StatusCode == 409 { // Conflict
				if attempt < maxRetries {
					// Wait briefly before retrying (exponential backoff)
					time.Sleep(time.Duration(attempt*50) * time.Millisecond)
					continue
				}
				// Max retries reached
				return nil, internal.NewConflictError("Booking failed due to concurrent reservations. Please try again.", err)
			}
		}

		// Non-conflict error, return immediately
		return nil, err
	}

	// Get additional info for confirmation (service name, child name)
	confirmation, err := s.buildBookingConfirmation(ctx, booking)
	if err != nil {
		// Booking created successfully but failed to build confirmation
		// Return basic confirmation
		return &services.BookingConfirmation{
			BookingID:     booking.ID,
			BookingNumber: booking.BookingNumber,
			BookingType:   booking.BookingType,
			TotalSessions: booking.TotalSessions,
			TotalAmount:   booking.TotalAmount,
			Status:        booking.Status,
			CreatedAt:     booking.CreatedAt,
		}, nil
	}

	return confirmation, nil
}

// buildBookingConfirmation builds a complete booking confirmation with all details
func (s *ServiceUsecase) buildBookingConfirmation(ctx context.Context, booking *services.Booking) (*services.BookingConfirmation, error) {
	// Fetch service name
	serviceName, err := s.repo.GetServiceNameByID(ctx, booking.ServiceID)
	if err != nil {
		// Log error but continue with empty service name
		serviceName = ""
	}

	// Fetch child name
	childName, err := s.repo.GetChildNameByID(ctx, booking.ChildID)
	if err != nil {
		// Log error but continue with empty child name
		childName = ""
	}

	// Build session info with coach names
	sessions := make([]*services.BookingSessionInfo, 0, len(booking.Sessions))
	for _, session := range booking.Sessions {
		var coachName *string
		if session.CoachID != nil {
			name, err := s.repo.GetCoachNameByID(ctx, *session.CoachID)
			if err == nil {
				coachName = &name
			}
		}

		sessions = append(sessions, &services.BookingSessionInfo{
			SessionDate: session.SessionDate,
			StartTime:   session.StartTime,
			EndTime:     session.EndTime,
			CoachName:   coachName,
		})
	}

	return &services.BookingConfirmation{
		BookingID:     booking.ID,
		BookingNumber: booking.BookingNumber,
		ServiceName:   serviceName,
		ChildName:     childName,
		BookingType:   booking.BookingType,
		TotalSessions: booking.TotalSessions,
		TotalAmount:   booking.TotalAmount,
		Sessions:      sessions,
		Status:        booking.Status,
		CreatedAt:     booking.CreatedAt,
	}, nil
}

// GetBooking retrieves a booking by ID
func (s *ServiceUsecase) GetBooking(ctx context.Context, bookingID int64, parentID int64) (*v1.BookingDetailResponse, error) {
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("Booking not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Verify the booking belongs to the parent
	if booking.ParentID != parentID {
		return nil, internal.NewForbiddenError("Access denied")
	}

	// Fetch enrichment data (service, child, vendor names)
	enrichment, err := s.repo.GetBookingEnrichment(ctx, booking.ServiceID, booking.ChildID, booking.VendorID)
	if err != nil {
		// Log error but continue with empty enrichment
		enrichment = &BookingEnrichment{}
	}

	// Convert to v1 response
	return ToV1BookingDetail(booking, enrichment), nil
}
