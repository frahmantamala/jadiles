package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

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

type LoginParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterParentParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	City     string `json:"city"`
	District string `json:"district"`
}

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

func (p *RegisterParentParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}

	return nil
}

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

func (p *RegisterVendorParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}

	return nil
}

func NewLoginParams(r *http.Request) (*LoginParams, error) {
	var params LoginParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return nil, err
	}
	return &params, nil
}

func (p *LoginParams) Validate(ctx context.Context) error {
	err := common.ValidateStruct(p)
	if err != nil {
		return err
	}

	return nil
}
