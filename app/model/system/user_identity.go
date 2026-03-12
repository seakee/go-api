// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UserIdentity represents a third-party identity bound to a local admin account.
type UserIdentity struct {
	gorm.Model

	UserID          uint       `gorm:"column:user_id" json:"user_id"`
	Provider        string     `gorm:"column:provider" json:"provider"`
	ProviderTenant  string     `gorm:"column:provider_tenant" json:"provider_tenant"`
	ProviderSubject string     `gorm:"column:provider_subject" json:"provider_subject"`
	DisplayName     string     `gorm:"column:display_name" json:"display_name"`
	AvatarURL       string     `gorm:"column:avatar_url" json:"avatar_url"`
	RawProfileJSON  string     `gorm:"column:raw_profile_json" json:"-"`
	BoundAt         *time.Time `gorm:"column:bound_at" json:"bound_at,omitempty"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`

	// Fields for Where support
	queryCondition string
	queryArgs      []interface{}
}

// Where sets query condition and arguments for the UserIdentity instance.
//
// Parameters:
//   - query: SQL query string or condition.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - *UserIdentity: pointer to the UserIdentity instance for method chaining.
func (u *UserIdentity) Where(query string, args ...interface{}) *UserIdentity {
	u.queryCondition = query
	u.queryArgs = args
	return u
}

// TableName specifies the table name for the UserIdentity model.
func (u *UserIdentity) TableName() string {
	return "sys_user_identity"
}

// First retrieves the first user identity matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *UserIdentity: pointer to the retrieved user identity, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (u *UserIdentity) First(ctx context.Context, db *gorm.DB) (*UserIdentity, error) {
	var identity UserIdentity

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.First(&identity).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &identity, nil
}

// Last retrieves the last user identity matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *UserIdentity: pointer to the retrieved user identity, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (u *UserIdentity) Last(ctx context.Context, db *gorm.DB) (*UserIdentity, error) {
	var identity UserIdentity

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.Order("id desc").First(&identity).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &identity, nil
}

// Create inserts a new user identity into the database and returns the ID of the created record.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - uint: ID of the created user identity.
//   - error: error if the insert operation fails, otherwise nil.
func (u *UserIdentity) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return u.ID, nil
}

// Delete soft-deletes the user identity from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - error: error if the delete operation fails, otherwise nil.
func (u *UserIdentity) Delete(ctx context.Context, db *gorm.DB) error {
	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database delete operation with context.
	return query.Delete(&UserIdentity{}).Error
}

// UnscopedDelete hard-deletes the user identity from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: number of rows affected by the delete.
//   - error: error if the delete operation fails, otherwise nil.
func (u *UserIdentity) UnscopedDelete(ctx context.Context, db *gorm.DB) (int64, error) {
	// Build query based on whether Where was used
	query := db.WithContext(ctx).Unscoped()
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database hard delete operation with context.
	result := query.Delete(&UserIdentity{})
	if result.Error != nil {
		return 0, fmt.Errorf("unscoped delete failed: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// Updates applies the specified updates to the user identity in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
//   - error: error if the update operation fails, otherwise nil.
func (u *UserIdentity) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query based on whether Where was used
	query := db.WithContext(ctx).Model(&UserIdentity{})
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database update operation with context.
	return query.Updates(updates).Error
}

// List retrieves all user identities matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - []UserIdentity: slice of retrieved user identities.
//   - error: error if the query fails, otherwise nil.
func (u *UserIdentity) List(ctx context.Context, db *gorm.DB) ([]UserIdentity, error) {
	var identities []UserIdentity

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.Find(&identities).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return identities, nil
}

// Count counts the number of user identities matching the criteria in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: count of matching user identities.
//   - error: error if the count operation fails, otherwise nil.
func (u *UserIdentity) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Build query based on whether Where was used
	query := db.WithContext(ctx).Model(&UserIdentity{})
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database count operation with context.
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// FindWithSort retrieves user identities matching the criteria from the database with sorting support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - sort: sorting criteria (e.g., "id desc").
//
// Returns:
//   - []UserIdentity: slice of retrieved user identities.
//   - error: error if the query fails, otherwise nil.
func (u *UserIdentity) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]UserIdentity, error) {
	var identities []UserIdentity

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&identities).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return identities, nil
}
