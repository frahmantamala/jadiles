package services

import (
	"fmt"
	"time"
)

// ================== Booking Domain Models ==================

// BookingType represents the type of booking
type BookingType string

const (
	BookingTypeTrial     BookingType = "trial"
	BookingTypeSingle    BookingType = "single"
	BookingTypePackage4  BookingType = "package_4"
	BookingTypePackage8  BookingType = "package_8"
	BookingTypePackage12 BookingType = "package_12"
)

// IsValid checks if booking type is valid
func (bt BookingType) IsValid() bool {
	switch bt {
	case BookingTypeTrial, BookingTypeSingle, BookingTypePackage4, BookingTypePackage8, BookingTypePackage12:
		return true
	}
	return false
}

// GetSessionCount returns number of sessions for the booking type
func (bt BookingType) GetSessionCount() int {
	switch bt {
	case BookingTypeTrial, BookingTypeSingle:
		return 1
	case BookingTypePackage4:
		return 4
	case BookingTypePackage8:
		return 8
	case BookingTypePackage12:
		return 12
	}
	return 0
}

// BookingStatus represents booking status
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
)

// Booking represents a booking domain model
type Booking struct {
	ID             int64
	BookingNumber  string
	ParentID       int64
	ChildID        int64
	ServiceID      int64
	VendorID       int64
	BookingType    BookingType
	TotalSessions  int
	TotalAmount    float64
	Status         BookingStatus
	PreferredCoach *int64
	ParentNotes    *string
	Version        int // For optimistic locking
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Related data (not stored in bookings table)
	Sessions []*BookingSession
}

// BookingSession represents a single session in a booking
type BookingSession struct {
	ID          int64
	BookingID   int64
	ScheduleID  int64
	SessionDate time.Time
	StartTime   string
	EndTime     string
	Status      SessionStatus
	CoachID     *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SessionStatus represents session status
type SessionStatus string

const (
	SessionStatusScheduled SessionStatus = "scheduled"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusNoShow    SessionStatus = "no_show"
)

// CreateBookingRequest represents a booking creation request
type CreateBookingRequest struct {
	ParentID       int64
	ChildID        int64
	ServiceID      int64
	BookingType    BookingType
	SessionDates   []BookingSessionRequest
	PreferredCoach *int64
	ParentNotes    *string
}

// BookingSessionRequest represents a session to be booked
type BookingSessionRequest struct {
	ScheduleID  int64
	SessionDate time.Time
}

// Validate validates the create booking request
func (r *CreateBookingRequest) Validate() error {
	if r.ParentID == 0 {
		return fmt.Errorf("parent_id is required")
	}
	if r.ChildID == 0 {
		return fmt.Errorf("child_id is required")
	}
	if r.ServiceID == 0 {
		return fmt.Errorf("service_id is required")
	}
	if !r.BookingType.IsValid() {
		return fmt.Errorf("invalid booking_type: %s", r.BookingType)
	}

	expectedSessions := r.BookingType.GetSessionCount()
	if len(r.SessionDates) != expectedSessions {
		return fmt.Errorf("booking_type %s requires %d sessions, got %d", r.BookingType, expectedSessions, len(r.SessionDates))
	}

	// Validate each session
	for i, session := range r.SessionDates {
		if session.ScheduleID == 0 {
			return fmt.Errorf("session %d: schedule_id is required", i+1)
		}
		if session.SessionDate.IsZero() {
			return fmt.Errorf("session %d: session_date is required", i+1)
		}
	}

	return nil
}

// BookingConfirmation represents booking confirmation response
type BookingConfirmation struct {
	BookingID     int64
	BookingNumber string
	ServiceName   string
	ChildName     string
	BookingType   BookingType
	TotalSessions int
	TotalAmount   float64
	Sessions      []*BookingSessionInfo
	Status        BookingStatus
	CreatedAt     time.Time
}

// BookingSessionInfo represents session info in confirmation
type BookingSessionInfo struct {
	SessionDate time.Time
	StartTime   string
	EndTime     string
	CoachName   *string
}
