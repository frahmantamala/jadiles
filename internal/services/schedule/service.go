package schedule

import (
	"context"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/services"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// Repository defines the data access interface for schedule capability
type Repository interface {
	GetSchedulesByService(ctx context.Context, serviceID int64) ([]*postgresql.ScheduleData, error)
	GetScheduleExceptions(ctx context.Context, serviceID int64, startDate, endDate time.Time) ([]*postgresql.ScheduleExceptionData, error)
	GetBookedSlotsCount(ctx context.Context, scheduleID int64, date time.Time) (int, error)
}

// ServiceUsecase handles schedule business logic
type ServiceUsecase struct {
	repo Repository
}

// NewService creates a new schedule service
func NewService(repo Repository) *ServiceUsecase {
	return &ServiceUsecase{
		repo: repo,
	}
}

// GetServiceAvailability retrieves monthly availability calendar
func (s *ServiceUsecase) GetServiceAvailability(ctx context.Context, serviceID int64, params *GetAvailabilityParams) (*v1.ServiceAvailabilityResponse, error) {
	// Parse month
	monthDate, err := time.Parse("2006-01", params.Month)
	if err != nil {
		return nil, internal.NewValidationError("month must be in YYYY-MM format")
	}
	year, month := monthDate.Year(), int(monthDate.Month())

	// Build monthly availability
	availability, err := s.buildMonthlyAvailability(ctx, serviceID, year, month)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to v1.ServiceAvailabilityResponse
	response := ToV1AvailabilityResponse(availability)

	return response, nil
}

// buildMonthlyAvailability builds the calendar view for a month (business logic)
func (s *ServiceUsecase) buildMonthlyAvailability(ctx context.Context, serviceID int64, year int, month int) ([]*services.DayAvailability, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of month

	// Fetch all schedules for this service
	schedules, err := s.repo.GetSchedulesByService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Fetch exceptions for this month
	exceptions, err := s.repo.GetScheduleExceptions(ctx, serviceID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Build calendar day by day
	var availability []*services.DayAvailability
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dayData := &services.DayAvailability{
			Date:  d,
			Slots: []*services.AvailabilitySlot{},
		}

		// Check for exception (holiday/closure)
		if exc := findException(exceptions, d); exc != nil {
			dayData.Exception = &services.ScheduleException{
				Date:     exc.ExceptionDate,
				Reason:   exc.Reason,
				IsClosed: exc.IsClosed,
			}
			availability = append(availability, dayData)
			continue
		}

		// Add slots for this day of week
		dow := int(d.Weekday())
		for _, schedule := range schedules {
			if schedule.DayOfWeek == dow {
				// Calculate available slots for this specific date
				availableSlots, err := s.repo.GetBookedSlotsCount(ctx, schedule.ID, d)
				if err != nil {
					return nil, err
				}

				dayData.Slots = append(dayData.Slots, &services.AvailabilitySlot{
					ScheduleID:     schedule.ID,
					StartTime:      schedule.StartTime,
					EndTime:        schedule.EndTime,
					AvailableSlots: availableSlots,
					CoachName:      schedule.CoachName,
				})
			}
		}

		availability = append(availability, dayData)
	}

	return availability, nil
}

// findException looks for an exception on a specific date
func findException(exceptions []*postgresql.ScheduleExceptionData, date time.Time) *postgresql.ScheduleExceptionData {
	for _, exc := range exceptions {
		if exc.ExceptionDate.Year() == date.Year() &&
			exc.ExceptionDate.Month() == date.Month() &&
			exc.ExceptionDate.Day() == date.Day() {
			return exc
		}
	}
	return nil
}
