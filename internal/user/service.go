package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	authpkg "github.com/frahmantamala/jadiles/internal/auth"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*datamodel.User, error)
	GetUserByID(ctx context.Context, id int64) (*datamodel.User, error)
	CreateUser(ctx context.Context, user *datamodel.User) error
	CreateParentProfile(ctx context.Context, profile *datamodel.ParentProfile) error
	CreateVendor(ctx context.Context, vendor *datamodel.Vendor) error
	// Transaction-based methods
	CreateParentWithProfile(ctx context.Context, user *datamodel.User, profile *datamodel.ParentProfile) error
	CreateVendorWithBusiness(ctx context.Context, user *datamodel.User, vendor *datamodel.Vendor) error
}

type TokenStorage interface {
	StoreToken(ctx context.Context, userID int64, token string, expiry time.Duration) error
	IsTokenValid(ctx context.Context, userID int64, token string) (bool, error)
	InvalidateToken(ctx context.Context, userID int64, token string) error
	InvalidateAllUserTokens(ctx context.Context, userID int64) error
}

type Service struct {
	repo            Repository
	jwtAuth         *authpkg.JWTAuthentication
	passwordManager *authpkg.PasswordManager
	tokenStorage    TokenStorage
}

func NewService(
	repo Repository,
	jwtAuth *authpkg.JWTAuthentication,
	passwordManager *authpkg.PasswordManager,
	tokenStorage TokenStorage,
) *Service {
	return &Service{
		repo:            repo,
		jwtAuth:         jwtAuth,
		passwordManager: passwordManager,
		tokenStorage:    tokenStorage,
	}
}

func (s *Service) RegisterParent(ctx context.Context, params *RegisterParentParams) (*v1.RegisterResponse, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, internal.NewInternalServerError(err)
	}
	if existingUser != nil {
		return nil, internal.ErrUserExist
	}

	domainUser := &User{
		Email:    params.Email,
		FullName: params.FullName,
		Phone:    params.Phone,
		Role:     RoleParent,
		Status:   StatusActive,
	}

	if err := domainUser.ValidatePassword(params.Password); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	if err := domainUser.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	hashedPassword, err := s.passwordManager.HashPassword(params.Password)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	userDM := params.RegParentToDataModel(hashedPassword)
	userDM.Version = 1

	domainProfile := &ParentProfile{
		City:     params.City,
		District: params.District,
	}

	if err := domainProfile.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	pp := &ParentProfileParams{
		City:     params.City,
		District: params.District,
	}
	profileDM := pp.PPToDataModel()
	profileDM.Version = 1

	if err := s.repo.CreateParentWithProfile(ctx, userDM, profileDM); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Build response
	resp := &v1.RegisterResponse{}
	resp.Data.User = v1.User{
		Id:        &userDM.ID,
		Email:     &userDM.Email,
		FullName:  &userDM.FullName,
		Phone:     &userDM.Phone,
		Role:      (*v1.UserRole)(&userDM.Role),
		Status:    (*v1.UserStatus)(&userDM.Status),
		CreatedAt: &userDM.CreatedAt,
	}

	return resp, nil
}

func (s *Service) RegisterVendor(ctx context.Context, params *RegisterVendorParams) (*v1.RegisterVendorResponse, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, internal.NewInternalServerError(err)
	}
	if existingUser != nil {
		return nil, internal.ErrUserExist
	}

	domainUser := &User{
		Email:    params.Email,
		FullName: params.FullName,
		Phone:    params.Phone,
		Role:     RoleVendor,
		Status:   StatusActive,
	}

	if err := domainUser.ValidatePassword(params.Password); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	if err := domainUser.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	domainVendor := &Vendor{
		BusinessName: params.BusinessName,
		BusinessType: params.BusinessType,
		Phone:        params.Phone,
		Whatsapp:     params.Whatsapp,
		Address:      params.Address,
		City:         params.City,
		District:     params.District,
		Status:       VendorStatusPending,
	}

	if err := domainVendor.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	hashedPassword, err := s.passwordManager.HashPassword(params.Password)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Prepare user data model with version for optimistic locking
	userDM := &datamodel.User{
		Email:        params.Email,
		PasswordHash: hashedPassword,
		FullName:     params.FullName,
		Phone:        params.Phone,
		Role:         "vendor",
		Status:       "active",
		Version:      1, // Initialize version for optimistic locking
	}

	// Prepare vendor data model with version for optimistic locking
	var whatsappPtr *string
	if params.Whatsapp != "" {
		w := params.Whatsapp
		whatsappPtr = &w
	}
	var districtPtr *string
	if params.District != "" {
		d := params.District
		districtPtr = &d
	}

	vendorDM := &datamodel.Vendor{
		BusinessName: params.BusinessName,
		BusinessType: params.BusinessType,
		Phone:        params.Phone,
		Whatsapp:     whatsappPtr,
		Address:      params.Address,
		City:         params.City,
		District:     districtPtr,
		Status:       "pending",
		Version:      1, // Initialize version for optimistic locking
	}

	// Create user and vendor in transaction (repository handles transaction)
	if err := s.repo.CreateVendorWithBusiness(ctx, userDM, vendorDM); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	resp := &v1.RegisterVendorResponse{}
	resp.Data.User = v1.User{
		Id:        &userDM.ID,
		Email:     &userDM.Email,
		FullName:  &userDM.FullName,
		Phone:     &userDM.Phone,
		Role:      (*v1.UserRole)(&userDM.Role),
		Status:    (*v1.UserStatus)(&userDM.Status),
		CreatedAt: &userDM.CreatedAt,
	}
	resp.Data.Vendor = v1.Vendor{
		Id:           &vendorDM.ID,
		BusinessName: &vendorDM.BusinessName,
		BusinessType: (*v1.BusinessType)(&vendorDM.BusinessType),
		Address:      &vendorDM.Address,
		City:         &vendorDM.City,
		Status:       (*v1.VendorStatus)(&vendorDM.Status),
	}

	message := "Vendor registration successful. Awaiting admin approval."
	resp.Message = &message
	return resp, nil
}

func (s *Service) Login(ctx context.Context, params *LoginParams) (*v1.LoginResponse, error) {
	userDM, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewUnauthorizedError("Invalid credentials")
		}
		return nil, internal.NewInternalServerError(err)
	}

	if userDM == nil {
		return nil, internal.NewUnauthorizedError("Invalid credentials")
	}

	// Verify password
	if err := s.passwordManager.VerifyPassword(userDM.PasswordHash, params.Password); err != nil {
		return nil, internal.NewUnauthorizedError("Invalid credentials")
	}

	// Create domain user to check login capability
	domainUser := &User{
		ID:     userDM.ID,
		Email:  userDM.Email,
		Status: UserStatus(userDM.Status),
	}

	// Check if user can login (domain rule)
	if err := domainUser.CanLogin(); err != nil {
		return nil, internal.NewForbiddenError(err.Error())
	}

	// Generate access token
	accessToken, _, err := s.jwtAuth.GenerateAccessToken(ctx, userDM.ID, userDM.Email, userDM.Role)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Store access token
	if err := s.tokenStorage.StoreToken(ctx, userDM.ID, accessToken, s.jwtAuth.AccessTokenDuration()); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Generate refresh token
	refreshToken, _, err := s.jwtAuth.GenerateRefreshToken(ctx, userDM.ID, userDM.Email, userDM.Role)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Store refresh token
	if err := s.tokenStorage.StoreToken(ctx, userDM.ID, refreshToken, s.jwtAuth.RefreshTokenDuration()); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Build response
	resp := &v1.LoginResponse{}
	resp.Data.Token = accessToken
	resp.Data.RefreshToken = refreshToken

	return resp, nil
}

func (s *Service) Logout(ctx context.Context, userID int64, token string) error {
	if s == nil || s.tokenStorage == nil {
		return internal.NewInternalServerError(errors.New("authentication service not initialized"))
	}
	// Invalidate the access token
	if err := s.tokenStorage.InvalidateToken(ctx, userID, token); err != nil {
		return internal.NewInternalServerError(err)
	}

	// Optionally: invalidate all user tokens for complete logout
	if err := s.tokenStorage.InvalidateAllUserTokens(ctx, userID); err != nil {
		return internal.NewInternalServerError(err)
	}

	return nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*v1.LoginResponse, error) {
	if s == nil || s.jwtAuth == nil || s.tokenStorage == nil || s.repo == nil {
		return nil, internal.NewInternalServerError(errors.New("authentication service not initialized"))
	}
	// Parse and validate refresh token
	claims, err := s.jwtAuth.ParseRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, internal.NewUnauthorizedError("Invalid refresh token")
	}

	// Check if refresh token is still valid in storage
	valid, err := s.tokenStorage.IsTokenValid(ctx, claims.UserID, refreshToken)
	if err != nil || !valid {
		return nil, internal.NewUnauthorizedError("Refresh token has been revoked")
	}

	// Get user from database to verify they still exist and are active
	userDM, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewUnauthorizedError("User not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Check if user can still login
	domainUser := &User{
		ID:     userDM.ID,
		Status: UserStatus(userDM.Status),
	}
	if err := domainUser.CanLogin(); err != nil {
		return nil, internal.NewForbiddenError(err.Error())
	}

	// Generate new access token
	accessToken, _, err := s.jwtAuth.GenerateAccessToken(ctx, userDM.ID, userDM.Email, userDM.Role)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Store new access token
	if err := s.tokenStorage.StoreToken(ctx, userDM.ID, accessToken, s.jwtAuth.AccessTokenDuration()); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Build response
	resp := &v1.LoginResponse{}
	resp.Data.Token = accessToken
	resp.Data.RefreshToken = refreshToken // Keep the same refresh token

	return resp, nil
}

func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	userDM, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("User not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to domain model
	domainUser := &User{
		ID:            userDM.ID,
		Email:         userDM.Email,
		FullName:      userDM.FullName,
		Phone:         userDM.Phone,
		Role:          UserRole(userDM.Role),
		Status:        UserStatus(userDM.Status),
		EmailVerified: userDM.EmailVerified,
		PhoneVerified: userDM.PhoneVerified,
		CreatedAt:     userDM.CreatedAt,
		UpdatedAt:     userDM.UpdatedAt,
	}

	return domainUser, nil
}
