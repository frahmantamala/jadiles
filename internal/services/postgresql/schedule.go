package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
)

// ScheduleData represents schedule data from database
type ScheduleData struct {
	ID             int64
	DayOfWeek      int
	StartTime      string
	EndTime        string
	AvailableSlots int
	CoachID        *int64
	CoachName      *string
	IsActive       bool
}

// ScheduleExceptionData represents exception from database
type ScheduleExceptionData struct {
	ID            int64
	ScheduleID    *int64
	ServiceID     *int64
	VendorID      *int64
	ExceptionDate time.Time
	Reason        *string
	IsClosed      bool
}

// GetSchedulesByService fetches all active schedules for a service
func (r *Repository) GetSchedulesByService(ctx context.Context, serviceID int64) ([]*ScheduleData, error) {
	var schedules []*ScheduleData
	query := `
		SELECT
			s.id, s.day_of_week, s.start_time, s.end_time, s.available_slots,
			s.coach_id, c.full_name as coach_name, s.is_active
		FROM schedules s
		LEFT JOIN coaches c ON s.coach_id = c.id
		WHERE s.service_id = $1 AND s.is_active = true
		ORDER BY s.day_of_week ASC, s.start_time ASC
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&schedules).Error
	return schedules, err
}

// GetScheduleExceptions fetches holidays and closures for a date range
func (r *Repository) GetScheduleExceptions(ctx context.Context, serviceID int64, startDate, endDate time.Time) ([]*ScheduleExceptionData, error) {
	var exceptions []*ScheduleExceptionData
	query := `
		SELECT id, schedule_id, service_id, vendor_id, exception_date, reason, is_closed
		FROM schedule_exceptions
		WHERE (service_id = $1 OR vendor_id = (SELECT vendor_id FROM services WHERE id = $1))
		  AND exception_date BETWEEN $2 AND $3
		ORDER BY exception_date ASC
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID, startDate, endDate).Scan(&exceptions).Error
	return exceptions, err
}

// GetBookedSlotsCount calculates booked slots for a specific schedule and date
func (r *Repository) GetBookedSlotsCount(ctx context.Context, scheduleID int64, date time.Time) (int, error) {
	var result struct {
		TotalSlots     int
		BookedCount    int
		AvailableSlots int
	}

	query := `
		SELECT
			s.available_slots as total_slots,
			COALESCE(COUNT(bs.id), 0) as booked_count,
			(s.available_slots - COALESCE(COUNT(bs.id), 0)) as available_slots
		FROM schedules s
		LEFT JOIN booking_sessions bs ON
			s.id = bs.schedule_id
			AND bs.session_date = $1
			AND bs.status NOT IN ('cancelled', 'no_show')
		WHERE s.id = $2
		GROUP BY s.id, s.available_slots
	`

	err := r.db.WithContext(ctx).Raw(query, date, scheduleID).Scan(&result).Error
	if err != nil {
		if err == sql.ErrNoRows {
			// No bookings for this slot yet, return full capacity
			var schedule datamodel.Schedule
			err = r.db.WithContext(ctx).Where("id = ?", scheduleID).First(&schedule).Error
			if err != nil {
				return 0, err
			}
			return schedule.AvailableSlots, nil
		}
		return 0, err
	}

	return result.AvailableSlots, nil
}
