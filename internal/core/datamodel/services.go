package datamodel

import "time"

// ServiceCategory represents the service_categories table
type ServiceCategory struct {
	ID           int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	Name         string    `db:"name"`
	Slug         string    `db:"slug"`
	Description  *string   `db:"description"`
	Icon         *string   `db:"icon"`
	ColorHex     *string   `db:"color_hex"`
	DisplayOrder int       `db:"display_order"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
}

// Service represents the services table
type Services struct {
	ID              int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	VendorID        int64     `db:"vendor_id"`
	CategoryID      int64     `db:"category_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	AgeMin          *int      `db:"age_min"`
	AgeMax          *int      `db:"age_max"`
	SkillLevel      string    `db:"skill_level"` // beginner, intermediate, advanced, all_levels
	ClassType       string    `db:"class_type"`  // private, small_group, large_group
	MaxParticipants *int      `db:"max_participants"`
	DurationMinutes int       `db:"duration_minutes"`
	PricePerSession float64   `db:"price_per_session"`
	TrialPrice      *float64  `db:"trial_price"`
	Package4Price   *float64  `db:"package_4_price"`
	Package8Price   *float64  `db:"package_8_price"`
	Package12Price  *float64  `db:"package_12_price"`
	Requirements    *string   `db:"requirements"`
	WhatWillLearn   *string   `db:"what_will_learn"`
	Status          string    `db:"status"` // active, inactive, draft
	IsFeatured      bool      `db:"is_featured"`
	Version         int       `db:"version" gorm:"default:1"` // Optimistic locking
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// Schedule represents the schedules table
type Schedule struct {
	ID             int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	ServiceID      int64     `db:"service_id"`
	CoachID        *int64    `db:"coach_id"`
	DayOfWeek      int       `db:"day_of_week"` // 0=Sunday, 6=Saturday
	StartTime      string    `db:"start_time"`  // HH:MM:SS format
	EndTime        string    `db:"end_time"`    // HH:MM:SS format
	AvailableSlots int       `db:"available_slots"`
	IsActive       bool      `db:"is_active"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// Review represents the reviews table
type Review struct {
	ID                int64      `db:"id" gorm:"primaryKey,autoIncrement"`
	BookingID         int64      `db:"booking_id"`
	ServiceID         int64      `db:"service_id"`
	ParentID          int64      `db:"parent_id"`
	Rating            int        `db:"rating"` // 1-5
	ReviewText        *string    `db:"review_text"`
	DidChildEnjoy     *bool      `db:"did_child_enjoy"`
	WouldRecommend    bool       `db:"would_recommend"`
	Photos            *string    `db:"photos"` // JSONB
	VendorResponse    *string    `db:"vendor_response"`
	VendorRespondedAt *time.Time `db:"vendor_responded_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

// ServiceCoach represents the service_coaches join table
type ServiceCoach struct {
	ID        int64 `db:"id" gorm:"primaryKey,autoIncrement"`
	ServiceID int64 `db:"service_id"`
	CoachID   int64 `db:"coach_id"`
	IsPrimary bool  `db:"is_primary"`
}

// ScheduleException represents the schedule_exceptions table
type ScheduleException struct {
	ID            int64      `db:"id" gorm:"primaryKey,autoIncrement"`
	ScheduleID    *int64     `db:"schedule_id"`
	ServiceID     *int64     `db:"service_id"`
	VendorID      *int64     `db:"vendor_id"`
	ExceptionDate time.Time  `db:"exception_date"`
	Reason        *string    `db:"reason"`
	IsClosed      bool       `db:"is_closed"`
	CreatedAt     time.Time  `db:"created_at"`
}

// Coach represents the coaches table
type Coach struct {
	ID               int64      `db:"id" gorm:"primaryKey,autoIncrement"`
	UserID           int64      `db:"user_id"`
	VendorID         int64      `db:"vendor_id"`
	FullName         string     `db:"full_name"`
	Bio              *string    `db:"bio"`
	ExperienceYears  int        `db:"experience_years"`
	Education        *string    `db:"education"`
	Certifications   *string    `db:"certifications"` // JSONB
	Specializations  *string    `db:"specializations"` // JSONB
	Photo            *string    `db:"photo"`
	IsFeatured       bool       `db:"is_featured"`
	Status           string     `db:"status"` // active, inactive
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}
