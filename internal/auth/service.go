// handle usecase for auth service
package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// Repository interface for database operations
type Repository interface {
	// User operations
	GetUserByEmail(ctx context.Context, email string) (*datamodel.User, error)
	GetUserByID(ctx context.Context, id int64) (*datamodel.User, error)
	CreateUser(ctx context.Context, user *datamodel.User) error

	// Parent profile operations
	CreateParentProfile(ctx context.Context, profile *datamodel.ParentProfile) error

	// Vendor operations
	CreateVendor(ctx context.Context, vendor *datamodel.Vendor) error
}

// TokenStorage interface for Redis operations
type TokenStorage interface {
	StoreToken(ctx context.Context, userID int64, token string, expiry time.Duration) error
	IsTokenValid(ctx context.Context, userID int64, token string) (bool, error)
	InvalidateToken(ctx context.Context, userID int64, token string) error
	InvalidateAllUserTokens(ctx context.Context, userID int64) error
}

// Service handles authentication business logic
type Service struct {
	repo            Repository
	jwtManager      *JWTManager
	passwordManager *PasswordManager
	tokenStorage    TokenStorage
}

// NewService creates a new auth service
func NewService(repo Repository, jwtManager *JWTManager, passwordManager *PasswordManager, tokenStorage TokenStorage) *Service {
	return &Service{
		repo:            repo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
		tokenStorage:    tokenStorage,
	}
}

// RegisterParent registers a new parent user
func (s *Service) RegisterParent(ctx context.Context, req v1.RegisterParentRequest) (*v1.RegisterResponse, error) {
	// 1. Validate input
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// 3. Hash password
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 4. Create user
	user := &datamodel.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Role:         "parent",
		Status:       "active",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	profile := &datamodel.ParentProfile{
		UserID: user.ID,
		City: func() string {
			if req.City != nil {
				return *req.City
			}
			return ""
		}(),
		District: req.District,
	}

	if err := s.repo.CreateParentProfile(ctx, profile); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 6. Generate JWT token
	token, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 7. Store token in Redis
	if err := s.tokenStorage.StoreToken(ctx, user.ID, token, s.jwtManager.accessExpiry); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 8. Build response using OpenAPI contract
	response := &v1.RegisterResponse{
		Data: struct {
			Token string  `json:"token"`
			User  v1.User `json:"user"`
		}{
			Token: token,
			User: v1.User{
				Id:        &user.ID,
				Email:     &user.Email,
				FullName:  &user.FullName,
				Phone:     &user.Phone,
				Role:      ptrUserRole(v1.UserRole(user.Role)),
				Status:    ptrUserStatus(v1.UserStatus(user.Status)),
				CreatedAt: &user.CreatedAt,
			},
		},
	}

	message := "Registration successful"
	response.Message = &message

	return response, nil
}

// RegisterVendor registers a new vendor user
func (s *Service) RegisterVendor(ctx context.Context, req v1.RegisterVendorRequest) (*v1.RegisterVendorResponse, error) {
	// 1. Validate input
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// 2. Check if email already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, internal.NewInternalServerError(err)
	}
	if existingUser != nil {
		return nil, internal.NewConflictError("Email already registered", internal.ErrAlreadyExists)
	}

	// 3. Hash password
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 4. Create user
	user := &datamodel.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Role:         "vendor",
		Status:       "active",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 5. Create vendor profile
	vendor := &datamodel.Vendor{
		UserID:       user.ID,
		BusinessName: req.BusinessName,
		BusinessType: string(req.BusinessType),
		Phone:        req.Phone,
		Whatsapp:     req.Whatsapp,
		Address:      req.Address,
		City:         req.City,
		District:     req.District,
		Status:       "pending", // Vendor needs admin approval
	}

	if err := s.repo.CreateVendor(ctx, vendor); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 6. Build response (no token for vendor until approved)
	vendorStatus := v1.VendorStatus("pending")
	response := &v1.RegisterVendorResponse{
		Data: struct {
			User   v1.User   `json:"user"`
			Vendor v1.Vendor `json:"vendor"`
		}{
			User: v1.User{
				Id:        &user.ID,
				Email:     &user.Email,
				FullName:  &user.FullName,
				Phone:     &user.Phone,
				Role:      ptrUserRole(v1.UserRole(user.Role)),
				Status:    ptrUserStatus(v1.UserStatus(user.Status)),
				CreatedAt: &user.CreatedAt,
			},
			Vendor: v1.Vendor{
				Id:           &vendor.ID,
				BusinessName: &vendor.BusinessName,
				BusinessType: ptrBusinessType(v1.BusinessType(vendor.BusinessType)),
				Address:      &vendor.Address,
				City:         &vendor.City,
				Status:       &vendorStatus,
			},
		},
	}

	message := "Vendor registration successful. Awaiting admin approval."
	response.Message = &message

	return response, nil
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req v1.LoginRequest) (*v1.LoginResponse, error) {
	// 1. Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewUnauthorizedError("Invalid credentials")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// 2. Verify password
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, internal.NewUnauthorizedError("Invalid credentials")
	}

	// 3. Check if user is active
	if user.Status != "active" {
		return nil, internal.NewForbiddenError("Account is suspended")
	}

	// 4. Generate JWT token
	token, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 5. Store token in Redis
	if err := s.tokenStorage.StoreToken(ctx, user.ID, token, s.jwtManager.accessExpiry); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// 6. Build response
	response := &v1.LoginResponse{
		Data: struct {
			Token string  `json:"token"`
			User  v1.User `json:"user"`
		}{
			Token: token,
			User: v1.User{
				Id:        &user.ID,
				Email:     &user.Email,
				FullName:  &user.FullName,
				Phone:     &user.Phone,
				Role:      ptrUserRole(v1.UserRole(user.Role)),
				Status:    ptrUserStatus(v1.UserStatus(user.Status)),
				CreatedAt: &user.CreatedAt,
			},
		},
	}

	return response, nil
}

// Logout invalidates a user's token
func (s *Service) Logout(ctx context.Context, userID int64, token string) error {
	// Invalidate token in Redis
	if err := s.tokenStorage.InvalidateToken(ctx, userID, token); err != nil {
		return internal.NewInternalServerError(err)
	}

	return nil
}

// ValidateToken checks if a token is valid (exists in Redis)
func (s *Service) ValidateToken(ctx context.Context, userID int64, token string) (bool, error) {
	return s.tokenStorage.IsTokenValid(ctx, userID, token)
}

// Helper functions to create pointers
func ptrUserRole(r v1.UserRole) *v1.UserRole {
	return &r
}

func ptrUserStatus(s v1.UserStatus) *v1.UserStatus {
	return &s
}

func ptrBusinessType(bt v1.BusinessType) *v1.BusinessType {
	return &bt
}
