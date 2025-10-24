// this is form datamodel for db persistance where represent table in db
package datamodel

import "time"

// User represents the users table in the database
type User struct {
	ID            int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	Email         string    `db:"email"`
	PasswordHash  string    `db:"password_hash"`
	FullName      string    `db:"full_name"`
	Phone         string    `db:"phone"`
	Role          string    `db:"role"`   // parent, vendor, coach, admin
	Status        string    `db:"status"` // active, suspended
	EmailVerified bool      `db:"email_verified"`
	PhoneVerified bool      `db:"phone_verified"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// ParentProfile represents the parent_profiles table
type ParentProfile struct {
	ID           int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	UserID       int64     `db:"user_id" gorm:"foreignKey:UserID"`
	Address      *string   `db:"address"`
	City         string    `db:"city"`
	District     *string   `db:"district"`
	PostalCode   *string   `db:"postal_code"`
	ProfileImage *string   `db:"profile_image"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// Vendor represents the vendors table
type Vendor struct {
	ID              int64      `db:"id" gorm:"primaryKey,autoIncrement"`
	UserID          int64      `db:"user_id" gorm:"foreignKey:UserID"`
	BusinessName    string     `db:"business_name"`
	Description     *string    `db:"description"`
	BusinessType    string     `db:"business_type"` // swimming_school, tutoring_center, art_studio, individual_coach
	Phone           string     `db:"phone"`
	Whatsapp        *string    `db:"whatsapp"`
	Address         string     `db:"address"`
	City            string     `db:"city"`
	District        *string    `db:"district"`
	PostalCode      *string    `db:"postal_code"`
	Latitude        *float64   `db:"latitude"`
	Longitude       *float64   `db:"longitude"`
	GoogleMapsURL   *string    `db:"google_maps_url"`
	Logo            *string    `db:"logo"`
	CoverImage      *string    `db:"cover_image"`
	Photos          *string    `db:"photos"`    // JSONB stored as string
	Amenities       *string    `db:"amenities"` // JSONB stored as string
	BusinessLicense *string    `db:"business_license"`
	Status          string     `db:"status"` // pending, active, suspended, rejected
	RejectionReason *string    `db:"rejection_reason"`
	RatingAvg       float64    `db:"rating_avg"`
	TotalReviews    int        `db:"total_reviews"`
	TotalBookings   int        `db:"total_bookings"`
	Verified        bool       `db:"verified"`
	VerifiedAt      *time.Time `db:"verified_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type Children struct {
	ID           int64     `db:"id" gorm:"primaryKey,autoIncrement"`
	UserID       int64     `db:"user_id" gorm:"foreignKey:UserID"`
	Name         string    `db:"name"`
	Nickname     string    `db:"nickname"`
	DOB          time.Time `db:"date_of_birth"`
	Age          int       `db:"age"`
	SpecialNeeds string    `db:"special_needs"`
	Gender       string    `db:"gender"`
	Photo        *string   `db:"photo"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
