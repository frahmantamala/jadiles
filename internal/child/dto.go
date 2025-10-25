package child

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// AddChildParams represents parameters for adding a child
type AddChildParams struct {
	Name         string `json:"name" validate:"required,min=2,max=100"`
	Nickname     string `json:"nickname" validate:"omitempty,max=50"`
	DateOfBirth  string `json:"date_of_birth" validate:"required"` // Format: YYYY-MM-DD
	Gender       string `json:"gender" validate:"required,oneof=male female"`
	SpecialNeeds string `json:"special_needs" validate:"omitempty,max=500"`
	Photo        string `json:"photo" validate:"omitempty,url"`
}

// UpdateChildParams represents parameters for updating a child
type UpdateChildParams struct {
	Name         string `json:"name" validate:"omitempty,min=2,max=100"`
	Nickname     string `json:"nickname" validate:"omitempty,max=50"`
	DateOfBirth  string `json:"date_of_birth" validate:"omitempty"` // Format: YYYY-MM-DD
	Gender       string `json:"gender" validate:"omitempty,oneof=male female"`
	SpecialNeeds string `json:"special_needs" validate:"omitempty,max=500"`
	Photo        string `json:"photo" validate:"omitempty,url"`
}

// NewAddChildParams creates AddChildParams from HTTP request
func NewAddChildParams(r *http.Request) (*AddChildParams, error) {
	var req v1.AddChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}

	params := &AddChildParams{
		Name:        req.Name,
		DateOfBirth: req.DateOfBirth.Format("2006-01-02"),
		Gender:      string(req.Gender),
	}

	if req.Nickname != nil {
		params.Nickname = *req.Nickname
	}
	if req.SpecialNeeds != nil {
		params.SpecialNeeds = *req.SpecialNeeds
	}

	return params, nil
}

// NewUpdateChildParams creates UpdateChildParams from HTTP request
func NewUpdateChildParams(r *http.Request) (*UpdateChildParams, error) {
	var req v1.UpdateChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, internal.NewValidationError("invalid request body")
	}

	params := &UpdateChildParams{}

	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Nickname != nil {
		params.Nickname = *req.Nickname
	}
	if req.SpecialNeeds != nil {
		params.SpecialNeeds = *req.SpecialNeeds
	}

	return params, nil
}

// Validate validates AddChildParams
func (p *AddChildParams) Validate(ctx context.Context) error {
	if err := common.ValidateStruct(p); err != nil {
		return err
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", p.DateOfBirth); err != nil {
		return internal.NewValidationError("date_of_birth must be in format YYYY-MM-DD")
	}

	return nil
}

// Validate validates UpdateChildParams
func (p *UpdateChildParams) Validate(ctx context.Context) error {
	if err := common.ValidateStruct(p); err != nil {
		return err
	}

	// Validate date format if provided
	if p.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", p.DateOfBirth); err != nil {
			return internal.NewValidationError("date_of_birth must be in format YYYY-MM-DD")
		}
	}

	return nil
}

// ToDataModel converts AddChildParams to datamodel.Child
func (p *AddChildParams) ToDataModel(parentID int64) (*datamodel.Children, error) {
	dob, err := time.Parse("2006-01-02", p.DateOfBirth)
	if err != nil {
		return nil, err
	}

	child := &datamodel.Children{
		ParentID:    parentID,
		Name:        p.Name,
		DateOfBirth: dob,
		Gender:      p.Gender,
		Version:     1,
	}

	if p.Nickname != "" {
		child.Nickname = &p.Nickname
	}
	if p.SpecialNeeds != "" {
		child.SpecialNeeds = &p.SpecialNeeds
	}
	if p.Photo != "" {
		child.Photo = &p.Photo
	}

	return child, nil
}

// ToV1Child converts domain Child to v1.Child
func ToV1Child(c *Child) v1.Child {
	age := c.CalculateAge()
	gender := v1.Gender(c.Gender)

	dob := openapi_types.Date{Time: c.DateOfBirth}
	child := v1.Child{
		Id:          &c.ID,
		Name:        &c.Name,
		DateOfBirth: &dob,
		Gender:      &gender,
		Age:         &age,
		CreatedAt:   &c.CreatedAt,
	}

	if c.Nickname != "" {
		child.Nickname = &c.Nickname
	}
	if c.SpecialNeeds != "" {
		child.SpecialNeeds = &c.SpecialNeeds
	}

	return child
}

// ToV1ChildFromDataModel converts datamodel.Child to v1.Child
func ToV1ChildFromDataModel(dm *datamodel.Children) v1.Child {
	// Calculate age
	now := time.Now()
	age := now.Year() - dm.DateOfBirth.Year()
	if now.YearDay() < dm.DateOfBirth.YearDay() {
		age--
	}

	gender := v1.Gender(dm.Gender)
	dob := openapi_types.Date{Time: dm.DateOfBirth}

	child := v1.Child{
		Id:          &dm.ID,
		Name:        &dm.Name,
		DateOfBirth: &dob,
		Gender:      &gender,
		Age:         &age,
		CreatedAt:   &dm.CreatedAt,
	}

	if dm.Nickname != nil {
		child.Nickname = dm.Nickname
	}
	if dm.SpecialNeeds != nil {
		child.SpecialNeeds = dm.SpecialNeeds
	}
	if dm.Photo != nil {
		child.Photo = dm.Photo
	}

	return child
}
