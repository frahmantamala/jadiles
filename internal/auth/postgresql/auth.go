// this is for postgresql auth implementation
package postgresql

import (
	"context"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
)

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*datamodel.User, error) {
	var user datamodel.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*datamodel.User, error) {
	var user datamodel.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, user *datamodel.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// CreateParentProfile creates a new parent profile
func (r *Repository) CreateParentProfile(ctx context.Context, profile *datamodel.ParentProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// CreateVendor creates a new vendor
func (r *Repository) CreateVendor(ctx context.Context, vendor *datamodel.Vendor) error {
	return r.db.WithContext(ctx).Create(vendor).Error
}
