package child

import (
	"context"
	"database/sql"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

type Repository interface {
	GetChildByID(ctx context.Context, id int64) (*datamodel.Children, error)
	GetChildrenByParentID(ctx context.Context, parentID int64) ([]*datamodel.Children, error)
	CreateChild(ctx context.Context, child *datamodel.Children) error
	UpdateChild(ctx context.Context, child *datamodel.Children) error
	DeleteChild(ctx context.Context, id int64, parentID int64) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// AddChild adds a new child for a parent
func (s *Service) AddChild(ctx context.Context, parentID int64, params *AddChildParams) (*v1.ChildResponse, error) {
	// Parse and validate date of birth
	dob, err := time.Parse("2006-01-02", params.DateOfBirth)
	if err != nil {
		return nil, internal.NewValidationError("invalid date_of_birth format, use YYYY-MM-DD")
	}

	// Create domain child for validation
	domainChild := &Child{
		ParentID:     parentID,
		Name:         params.Name,
		Nickname:     params.Nickname,
		DateOfBirth:  dob,
		Gender:       Gender(params.Gender),
		SpecialNeeds: params.SpecialNeeds,
		Photo:        params.Photo,
	}

	// Validate domain rules
	if err := domainChild.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	// Convert to data model
	childDM, err := params.ToDataModel(parentID)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Create child
	if err := s.repo.CreateChild(ctx, childDM); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Build response
	resp := &v1.ChildResponse{}
	resp.Data = ToV1ChildFromDataModel(childDM)

	return resp, nil
}

// GetChildren retrieves all children for a parent
func (s *Service) GetChildren(ctx context.Context, parentID int64) (*v1.ChildrenListResponse, error) {
	childrenDM, err := s.repo.GetChildrenByParentID(ctx, parentID)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to response
	resp := &v1.ChildrenListResponse{}
	resp.Data = make([]v1.Child, len(childrenDM))

	for i, childDM := range childrenDM {
		resp.Data[i] = ToV1ChildFromDataModel(childDM)
	}

	return resp, nil
}

// GetChild retrieves a specific child
func (s *Service) GetChild(ctx context.Context, childID int64, parentID int64) (*v1.ChildResponse, error) {
	childDM, err := s.repo.GetChildByID(ctx, childID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("Child not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Verify parent ownership
	if childDM.ParentID != parentID {
		return nil, internal.NewForbiddenError("You don't have permission to access this child")
	}

	// Build response
	resp := &v1.ChildResponse{}
	resp.Data = ToV1ChildFromDataModel(childDM)

	return resp, nil
}

// UpdateChild updates a child's information
func (s *Service) UpdateChild(ctx context.Context, childID int64, parentID int64, params *UpdateChildParams) (*v1.ChildResponse, error) {
	// Get existing child
	childDM, err := s.repo.GetChildByID(ctx, childID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("Child not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Verify parent ownership
	if childDM.ParentID != parentID {
		return nil, internal.NewForbiddenError("You don't have permission to update this child")
	}

	// Update fields if provided
	if params.Name != "" {
		childDM.Name = params.Name
	}
	if params.Nickname != "" {
		nickname := params.Nickname
		childDM.Nickname = &nickname
	}
	if params.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", params.DateOfBirth)
		if err != nil {
			return nil, internal.NewValidationError("invalid date_of_birth format, use YYYY-MM-DD")
		}
		childDM.DateOfBirth = dob
	}
	if params.Gender != "" {
		childDM.Gender = params.Gender
	}
	if params.SpecialNeeds != "" {
		specialNeeds := params.SpecialNeeds
		childDM.SpecialNeeds = &specialNeeds
	}
	if params.Photo != "" {
		photo := params.Photo
		childDM.Photo = &photo
	}

	childDM.UpdatedAt = time.Now()

	// Create domain child for validation
	domainChild := &Child{
		ID:          childDM.ID,
		ParentID:    childDM.ParentID,
		Name:        childDM.Name,
		DateOfBirth: childDM.DateOfBirth,
		Gender:      Gender(childDM.Gender),
		Version:     childDM.Version,
	}

	if childDM.Nickname != nil {
		domainChild.Nickname = *childDM.Nickname
	}
	if childDM.SpecialNeeds != nil {
		domainChild.SpecialNeeds = *childDM.SpecialNeeds
	}
	if childDM.Photo != nil {
		domainChild.Photo = *childDM.Photo
	}

	// Validate domain rules
	if err := domainChild.Validate(); err != nil {
		return nil, internal.NewValidationError(err.Error())
	}

	// Update child
	if err := s.repo.UpdateChild(ctx, childDM); err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Build response
	resp := &v1.ChildResponse{}
	resp.Data = ToV1ChildFromDataModel(childDM)

	return resp, nil
}

// DeleteChild deletes a child
func (s *Service) DeleteChild(ctx context.Context, childID int64, parentID int64) error {
	// Verify ownership by attempting to get the child first
	childDM, err := s.repo.GetChildByID(ctx, childID)
	if err != nil {
		if err == sql.ErrNoRows {
			return internal.NewNotFoundError("Child not found")
		}
		return internal.NewInternalServerError(err)
	}

	// Verify parent ownership
	if childDM.ParentID != parentID {
		return internal.NewForbiddenError("You don't have permission to delete this child")
	}

	// Delete child
	if err := s.repo.DeleteChild(ctx, childID, parentID); err != nil {
		return internal.NewInternalServerError(err)
	}

	return nil
}
