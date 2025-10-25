package datamodel

import "time"

// Booking represents the bookings table
type Booking struct {
	ID             int64      `db:"id" gorm:"primaryKey,autoIncrement"`
	BookingNumber  string     `db:"booking_number" gorm:"uniqueIndex"`
	ParentID       int64      `db:"parent_id"`
	ChildID        int64      `db:"child_id"`
	ServiceID      int64      `db:"service_id"`
	VendorID       int64      `db:"vendor_id"`
	BookingType    string     `db:"booking_type"` // trial, single, package_4, package_8, package_12
	TotalSessions  int        `db:"total_sessions"`
	TotalAmount    float64    `db:"total_amount"`
	Status         string     `db:"status"` // pending, confirmed, cancelled, completed
	PreferredCoach *int64     `db:"preferred_coach"`
	ParentNotes    *string    `db:"parent_notes"`
	Version        int        `db:"version" gorm:"default:1"` // Optimistic locking
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

// TableName specifies the table name
func (Booking) TableName() string {
	return "bookings"
}

// BookingSession represents the booking_sessions table
type BookingSession struct {
	ID          int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	BookingID   int64     `db:"booking_id"`
	ScheduleID  int64     `db:"schedule_id"`
	SessionDate time.Time `db:"session_date"`
	Status      string    `db:"status"` // scheduled, completed, cancelled, no_show
	CoachID     *int64    `db:"coach_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// TableName specifies the table name
func (BookingSession) TableName() string {
	return "booking_sessions"
}
