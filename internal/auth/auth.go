package auth

import (
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User domain model
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
)

// ParentProfile domain model
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

// Vendor domain model
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

// Domain-specific errors are handled by internal package error constructors
// Use internal.NewConflictError(), internal.NewValidationError(), etc.

// JWT Claims structure
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken generates a new JWT access token
func (j *JWTManager) GenerateAccessToken(userID int64, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, internal.ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, internal.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, internal.ErrInvalidToken
	}

	return claims, nil
}

// PasswordManager handles password hashing and verification
type PasswordManager struct {
	cost int
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: bcrypt.DefaultCost, // cost 12
	}
}

// HashPassword hashes a plain text password
func (p *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies if a password matches the hash
func (p *PasswordManager) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) *internal.AppError {
	if len(email) < 3 || len(email) > 255 {
		return internal.NewAppError("VALIDATION_ERROR", "Invalid email format", 400, internal.ErrValidation)
	}
	// Add more validation if needed (regex, etc.)
	return nil
}

// ValidatePassword performs password strength validation
func ValidatePassword(password string) *internal.AppError {
	if len(password) < 8 || len(password) > 72 {
		return internal.NewAppError("VALIDATION_ERROR", "Invalid password format", 400, internal.ErrValidation)
	}
	return nil
}
