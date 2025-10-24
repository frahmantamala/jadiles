package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// User represents the core user domain model
type User struct {
	ID            int64
	Email         string
	PasswordHash  string
	FullName      string
	Phone         string
	Role          UserRole
	Status        UserStatus
	EmailVerified bool
	PhoneVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserRole string

const (
	RoleParent UserRole = "parent"
	RoleVendor UserRole = "vendor"
	RoleCoach  UserRole = "coach"
	RoleAdmin  UserRole = "admin"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusInactive  UserStatus = "inactive"
	StatusPending   UserStatus = "pending"
)

// ParentProfile represents parent-specific profile
type ParentProfile struct {
	ID           int64
	UserID       int64
	Address      string
	City         string
	District     string
	PostalCode   string
	ProfileImage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Vendor represents vendor domain model
type Vendor struct {
	ID              int64
	UserID          int64
	BusinessName    string
	Description     string
	BusinessType    string
	Phone           string
	Whatsapp        string
	Address         string
	City            string
	District        string
	PostalCode      string
	Latitude        float64
	Longitude       float64
	GoogleMapsURL   string
	Logo            string
	CoverImage      string
	Photos          []string
	Amenities       []string
	BusinessLicense string
	Status          VendorStatus
	RejectionReason string
	RatingAvg       float64
	TotalReviews    int
	TotalBookings   int
	Verified        bool
	VerifiedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type VendorStatus string

const (
	VendorStatusPending   VendorStatus = "pending"
	VendorStatusActive    VendorStatus = "active"
	VendorStatusSuspended VendorStatus = "suspended"
	VendorStatusRejected  VendorStatus = "rejected"
)

// Domain validation rules

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^(\+62|62|0)[0-9]{9,13}$`)
)

// ValidateEmail validates email format
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if len(u.Email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePhone validates phone number format
func (u *User) ValidatePhone() error {
	if u.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	// Remove spaces and dashes
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(u.Phone, " ", ""), "-", "")
	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone format, must be Indonesian phone number")
	}
	return nil
}

// ValidateFullName validates full name
func (u *User) ValidateFullName() error {
	if u.FullName == "" {
		return fmt.Errorf("full name is required")
	}
	if len(u.FullName) < 2 {
		return fmt.Errorf("full name must be at least 2 characters")
	}
	if len(u.FullName) > 255 {
		return fmt.Errorf("full name must not exceed 255 characters")
	}
	return nil
}

// ValidatePassword validates password strength (used before hashing)
func (u *User) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
	}
	return nil
}

// ValidateRole validates user role
func (u *User) ValidateRole() error {
	validRoles := []UserRole{RoleParent, RoleVendor, RoleCoach, RoleAdmin}
	for _, r := range validRoles {
		if u.Role == r {
			return nil
		}
	}
	return fmt.Errorf("invalid role: %s", u.Role)
}

// ValidateStatus validates user status
func (u *User) ValidateStatus() error {
	validStatuses := []UserStatus{StatusActive, StatusSuspended, StatusInactive, StatusPending}
	for _, s := range validStatuses {
		if u.Status == s {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s", u.Status)
}

// Validate validates all user fields
func (u *User) Validate() error {
	if err := u.ValidateEmail(); err != nil {
		return err
	}
	if err := u.ValidatePhone(); err != nil {
		return err
	}
	if err := u.ValidateFullName(); err != nil {
		return err
	}
	if err := u.ValidateRole(); err != nil {
		return err
	}
	if err := u.ValidateStatus(); err != nil {
		return err
	}
	return nil
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsSuspended checks if user is suspended
func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended
}

// CanLogin checks if user can login
func (u *User) CanLogin() error {
	if !u.IsActive() {
		if u.IsSuspended() {
			return fmt.Errorf("account is suspended")
		}
		return fmt.Errorf("account is not active")
	}
	return nil
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}

// ParentProfile domain methods

// ValidateCity validates city field
func (p *ParentProfile) ValidateCity() error {
	if p.City != "" && len(p.City) > 100 {
		return fmt.Errorf("city must not exceed 100 characters")
	}
	return nil
}

// ValidateDistrict validates district field
func (p *ParentProfile) ValidateDistrict() error {
	if p.District != "" && len(p.District) > 100 {
		return fmt.Errorf("district must not exceed 100 characters")
	}
	return nil
}

// Validate validates parent profile
func (p *ParentProfile) Validate() error {
	if p.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if err := p.ValidateCity(); err != nil {
		return err
	}
	if err := p.ValidateDistrict(); err != nil {
		return err
	}
	return nil
}

// Vendor domain methods

// ValidateBusinessName validates business name
func (v *Vendor) ValidateBusinessName() error {
	if v.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if len(v.BusinessName) < 3 {
		return fmt.Errorf("business name must be at least 3 characters")
	}
	if len(v.BusinessName) > 255 {
		return fmt.Errorf("business name must not exceed 255 characters")
	}
	return nil
}

// ValidateBusinessType validates business type
func (v *Vendor) ValidateBusinessType() error {
	validTypes := []string{"swimming_school", "tutoring_center", "art_studio", "individual_coach"}
	for _, t := range validTypes {
		if v.BusinessType == t {
			return nil
		}
	}
	return fmt.Errorf("invalid business type: %s", v.BusinessType)
}

// ValidateAddress validates address
func (v *Vendor) ValidateAddress() error {
	if v.Address == "" {
		return fmt.Errorf("address is required")
	}
	if len(v.Address) < 10 {
		return fmt.Errorf("address must be at least 10 characters")
	}
	if len(v.Address) > 500 {
		return fmt.Errorf("address must not exceed 500 characters")
	}
	return nil
}

// ValidateCity validates city
func (v *Vendor) ValidateCity() error {
	if v.City == "" {
		return fmt.Errorf("city is required")
	}
	if len(v.City) > 100 {
		return fmt.Errorf("city must not exceed 100 characters")
	}
	return nil
}

// ValidatePhone validates vendor phone
func (v *Vendor) ValidatePhone() error {
	if v.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(v.Phone, " ", ""), "-", "")
	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

// ValidateWhatsapp validates whatsapp number if provided
func (v *Vendor) ValidateWhatsapp() error {
	if v.Whatsapp != "" {
		cleanWhatsapp := strings.ReplaceAll(strings.ReplaceAll(v.Whatsapp, " ", ""), "-", "")
		if !phoneRegex.MatchString(cleanWhatsapp) {
			return fmt.Errorf("invalid whatsapp format")
		}
	}
	return nil
}

// ValidateStatus validates vendor status
func (v *Vendor) ValidateStatus() error {
	validStatuses := []VendorStatus{VendorStatusPending, VendorStatusActive, VendorStatusSuspended, VendorStatusRejected}
	for _, s := range validStatuses {
		if v.Status == s {
			return nil
		}
	}
	return fmt.Errorf("invalid vendor status: %s", v.Status)
}

// Validate validates all vendor fields
func (v *Vendor) Validate() error {
	if v.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if err := v.ValidateBusinessName(); err != nil {
		return err
	}
	if err := v.ValidateBusinessType(); err != nil {
		return err
	}
	if err := v.ValidateAddress(); err != nil {
		return err
	}
	if err := v.ValidateCity(); err != nil {
		return err
	}
	if err := v.ValidatePhone(); err != nil {
		return err
	}
	if err := v.ValidateWhatsapp(); err != nil {
		return err
	}
	if err := v.ValidateStatus(); err != nil {
		return err
	}
	return nil
}

// IsPending checks if vendor is pending approval
func (v *Vendor) IsPending() bool {
	return v.Status == VendorStatusPending
}

// IsActive checks if vendor is active
func (v *Vendor) IsActive() bool {
	return v.Status == VendorStatusActive
}

// IsRejected checks if vendor is rejected
func (v *Vendor) IsRejected() bool {
	return v.Status == VendorStatusRejected
}

// Approve approves the vendor
func (v *Vendor) Approve() {
	v.Status = VendorStatusActive
	v.Verified = true
	now := time.Now()
	v.VerifiedAt = &now
	v.UpdatedAt = now
	v.RejectionReason = ""
}

// Reject rejects the vendor with reason
func (v *Vendor) Reject(reason string) error {
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}
	v.Status = VendorStatusRejected
	v.RejectionReason = reason
	v.Verified = false
	v.VerifiedAt = nil
	v.UpdatedAt = time.Now()
	return nil
}

// Suspend suspends the vendor
func (v *Vendor) Suspend(reason string) {
	v.Status = VendorStatusSuspended
	v.RejectionReason = reason
	v.UpdatedAt = time.Now()
}

// UpdateRating updates vendor rating
func (v *Vendor) UpdateRating(newRating float64) error {
	if newRating < 0 || newRating > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}

	// Calculate new average
	totalRating := v.RatingAvg * float64(v.TotalReviews)
	v.TotalReviews++
	v.RatingAvg = (totalRating + newRating) / float64(v.TotalReviews)
	v.UpdatedAt = time.Now()

	return nil
}

// IncrementBooking increments total bookings
func (v *Vendor) IncrementBooking() {
	v.TotalBookings++
	v.UpdatedAt = time.Now()
}
