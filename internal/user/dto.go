package user

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// RegisterVendorParams represents vendor registration parameters
type RegisterVendorParams struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FullName     string `json:"full_name" validate:"required"`
	Phone        string `json:"phone" validate:"required"`
	BusinessName string `json:"business_name" validate:"required"`
	BusinessType string `json:"business_type" validate:"required,oneof=swimming_school tutoring_center art_studio individual_coach"`
	Address      string `json:"address" validate:"required"`
	City         string `json:"city" validate:"required"`
	District     string `json:"district"`
	Whatsapp     string `json:"whatsapp"`
}

// LoginParams represents login parameters
type LoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterParentParams represents parent registration parameters
type RegisterParentParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	City     string `json:"city"`
	District string `json:"district"`
}

// RefreshTokenParams represents refresh token parameters
type RefreshTokenParams struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ParentProfileParams represents parent profile parameters
type ParentProfileParams struct {
	UserDetails int64
	City        string
	District    string
}

// NewRegisterParentParams creates RegisterParentParams from HTTP request
func NewRegisterParentParams(r *http.Request) (*RegisterParentParams, error) {
	var req v1.RegisterParentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}

	params := &RegisterParentParams{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		City: func() string {
			if req.City != nil {
				return *req.City
			}
			return ""
		}(),
		District: func() string {
			if req.District != nil {
				return *req.District
			}
			return ""
		}(),
	}

	return params, nil
}

// RegParentToDataModel converts RegisterParentParams to datamodel.User
func (r *RegisterParentParams) RegParentToDataModel(hashPass string) *datamodel.User {
	userDataModel := &datamodel.User{
		Email:        r.Email,
		PasswordHash: hashPass,
		FullName:     r.FullName,
		Phone:        r.Phone,
		Role:         "parent",
		Status:       "active",
	}
	return userDataModel
}

// Validate validates RegisterParentParams
func (p *RegisterParentParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}
	return nil
}

// PPToDataModel converts ParentProfileParams to datamodel.ParentProfile
func (pp *ParentProfileParams) PPToDataModel() *datamodel.ParentProfile {
	parentProfile := &datamodel.ParentProfile{
		UserID:   pp.UserDetails,
		City:     pp.City,
		District: &pp.District,
	}
	return parentProfile
}

// NewRegisterVendorParams creates RegisterVendorParams from HTTP request
func NewRegisterVendorParams(r *http.Request) (*RegisterVendorParams, error) {
	var req v1.RegisterVendorRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}

	params := &RegisterVendorParams{
		Email:        req.Email,
		Password:     req.Password,
		FullName:     req.FullName,
		Phone:        req.Phone,
		BusinessName: req.BusinessName,
		BusinessType: string(req.BusinessType),
		Address:      req.Address,
		City:         req.City,
		District: func() string {
			if req.District != nil {
				return *req.District
			}
			return ""
		}(),
		Whatsapp: func() string {
			if req.Whatsapp != nil {
				return *req.Whatsapp
			}
			return ""
		}(),
	}

	return params, nil
}

// Validate validates RegisterVendorParams
func (p *RegisterVendorParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}
	return nil
}

// NewLoginParams creates LoginParams from HTTP request
func NewLoginParams(r *http.Request) (*LoginParams, error) {
	var params LoginParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}
	return &params, nil
}

// Validate validates LoginParams
func (p *LoginParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}
	return nil
}

// NewRefreshTokenParams creates RefreshTokenParams from HTTP request
func NewRefreshTokenParams(r *http.Request) (*RefreshTokenParams, error) {
	var params RefreshTokenParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}
	return &params, nil
}

// Validate validates RefreshTokenParams
func (p *RefreshTokenParams) Validate(ctx context.Context) error {
	if p.RefreshToken == "" {
		return internal.NewValidationError("refresh_token is required")
	}
	return nil
}

// ToV1User converts domain User to v1.User
func ToV1User(u *User) v1.User {
	return v1.User{
		Id:        &u.ID,
		Email:     &u.Email,
		FullName:  &u.FullName,
		Phone:     &u.Phone,
		Role:      (*v1.UserRole)(&u.Role),
		Status:    (*v1.UserStatus)(&u.Status),
		CreatedAt: &u.CreatedAt,
	}
}

// ToV1Vendor converts domain Vendor to v1.Vendor
func ToV1Vendor(v *Vendor) v1.Vendor {
	status := v1.VendorStatus(v.Status)
	businessType := v1.BusinessType(v.BusinessType)

	return v1.Vendor{
		Id:           &v.ID,
		BusinessName: &v.BusinessName,
		BusinessType: &businessType,
		Address:      &v.Address,
		City:         &v.City,
		Status:       &status,
	}
}
