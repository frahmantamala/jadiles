package postgresql

import (
	"context"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	"gorm.io/gorm"
)

// GetChildByID retrieves a child by ID
func (r *Repository) GetChildByID(ctx context.Context, id int64) (*datamodel.Children, error) {
	var child datamodel.Children
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&child).Error

	if err != nil {
		return nil, err
	}

	return &child, nil
}

// GetChildrenByParentID retrieves all children for a parent
func (r *Repository) GetChildrenByParentID(ctx context.Context, parentID int64) ([]*datamodel.Children, error) {
	var children []*datamodel.Children
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("created_at DESC").
		Find(&children).Error

	if err != nil {
		return nil, err
	}

	return children, nil
}

// CreateChild creates a new child
func (r *Repository) CreateChild(ctx context.Context, child *datamodel.Children) error {
	return r.db.WithContext(ctx).Create(child).Error
}

// UpdateChild updates a child with optimistic locking
func (r *Repository) UpdateChild(ctx context.Context, child *datamodel.Children) error {
	result := r.db.WithContext(ctx).
		Model(child).
		Where("id = ? AND version = ?", child.ID, child.Version).
		Updates(map[string]interface{}{
			"name":          child.Name,
			"nickname":      child.Nickname,
			"date_of_birth": child.DateOfBirth,
			"gender":        child.Gender,
			"special_needs": child.SpecialNeeds,
			"photo":         child.Photo,
			"version":       child.Version + 1,
			"updated_at":    child.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	child.Version++
	return nil
}

// DeleteChild soft deletes a child
func (r *Repository) DeleteChild(ctx context.Context, id int64, parentID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND parent_id = ?", id, parentID).
		Delete(&datamodel.Children{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
