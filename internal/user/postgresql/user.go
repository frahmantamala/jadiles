package postgresql

import (
	"context"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	"gorm.io/gorm"
)

// GetUserByEmail retrieves user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*datamodel.User, error) {
	var user datamodel.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves user by ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*datamodel.User, error) {
	var user datamodel.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

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

// CreateParentWithProfile creates a user and parent profile in a transaction
func (r *Repository) CreateParentWithProfile(ctx context.Context, user *datamodel.User, profile *datamodel.ParentProfile) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Set user_id in profile
		profile.UserID = user.ID

		// Create parent profile
		if err := tx.Create(profile).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateVendorWithBusiness creates a user and vendor in a transaction
func (r *Repository) CreateVendorWithBusiness(ctx context.Context, user *datamodel.User, vendor *datamodel.Vendor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Set user_id in vendor
		vendor.UserID = user.ID

		// Create vendor
		if err := tx.Create(vendor).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetParentProfileByUserID retrieves parent profile by user ID
func (r *Repository) GetParentProfileByUserID(ctx context.Context, userID int64) (*datamodel.ParentProfile, error) {
	var profile datamodel.ParentProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetVendorByUserID retrieves vendor by user ID
func (r *Repository) GetVendorByUserID(ctx context.Context, userID int64) (*datamodel.Vendor, error) {
	var vendor datamodel.Vendor
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&vendor).Error

	if err != nil {
		return nil, err
	}

	return &vendor, nil
}

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&datamodel.User{}, id).Error
}
