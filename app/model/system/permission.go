// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package system

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Permission represents a permission entry in the system access control.
type Permission struct {
	gorm.Model

	Name        string `gorm:"column:name" json:"name"`
	Type        string `gorm:"column:type" json:"type"`
	Method      string `gorm:"column:method" json:"method"`
	Path        string `gorm:"column:path" json:"path"`
	Description string `gorm:"column:description" json:"description"`
	Group       string `gorm:"column:group" json:"group"`

	// Query condition fields for Where method
	queryCondition string        `gorm:"-" json:"-"`
	queryArgs      []interface{} `gorm:"-" json:"-"`
}

// TableName specifies the table name for the Permission model.
func (p *Permission) TableName() string {
	return "sys_permission"
}

// Where sets query conditions for the Permission model.
//
// Parameters:
//   - condition: string SQL condition with placeholders.
//   - args: ...interface{} arguments to replace placeholders in the condition.
//
// Returns:
//   - *Permission: pointer to the Permission instance for method chaining.
func (p *Permission) Where(condition string, args ...interface{}) *Permission {
	p.queryCondition = condition
	p.queryArgs = args
	return p
}

// First retrieves the first permission matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Permission: pointer to the retrieved permission, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) First(ctx context.Context, db *gorm.DB) (*Permission, error) {
	var permission Permission

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database query
	if err := query.First(&permission).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &permission, nil
}

// Last retrieves the last permission matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Permission: pointer to the retrieved permission, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) Last(ctx context.Context, db *gorm.DB) (*Permission, error) {
	var permission Permission

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database query
	if err := query.Order("id desc").First(&permission).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &permission, nil
}

// Create inserts a new permission into the database and returns the ID of the created Permission.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - uint: ID of the created permission.
//   - error: error if the insert operation fails, otherwise nil.
func (p *Permission) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(p).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return p.ID, nil
}

// Delete removes the permission from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - error: error if the delete operation fails, otherwise nil.
func (p *Permission) Delete(ctx context.Context, db *gorm.DB) error {
	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		return query.Where(p.queryCondition, p.queryArgs...).Delete(&Permission{}).Error
	} else {
		return query.Where(p).Delete(&Permission{}).Error
	}
}

// Updates applies the specified updates to the permission in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
//   - error: error if the update operation fails, otherwise nil.
func (p *Permission) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query with context and explicitly set the model
	query := db.WithContext(ctx).Model(&Permission{})

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		return query.Where(p.queryCondition, p.queryArgs...).Updates(updates).Error
	} else {
		return query.Where(p).Updates(updates).Error
	}
}

// List retrieves all permissions matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - []Permission: slice of retrieved permissions.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) List(ctx context.Context, db *gorm.DB) ([]Permission, error) {
	var permissions []Permission

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database query
	if err := query.Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return permissions, nil
}

// ListByArgs retrieves permissions matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - []Permission: slice of retrieved permissions.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]Permission, error) {
	var permissions []Permission

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return permissions, nil
}

// CountByArgs counts the number of permissions matching the specified query and arguments in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - int64: count of matching permissions.
//   - error: error if the count operation fails, otherwise nil.
func (p *Permission) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&Permission{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of permissions matching the criteria in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: count of matching permissions.
//   - error: error if the count operation fails, otherwise nil.
func (p *Permission) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Build query with context
	query := db.WithContext(ctx).Model(&Permission{})

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database count operation
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// BatchInsert inserts multiple permissions into the database in a single batch operation.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - permissions: slice of Permission instances to be inserted.
//
// Returns:
//   - error: error if the batch insert operation fails, otherwise nil.
func (p *Permission) BatchInsert(ctx context.Context, db *gorm.DB, permissions []Permission) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&permissions).Error
}

// FindWithPagination retrieves permissions matching the criteria from the database with pagination support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - page: page number for pagination (1-based).
//   - size: number of permissions per page.
//
// Returns:
//   - []Permission: slice of retrieved permissions.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]Permission, error) {
	var permissions []Permission

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database query with context, applying offset and limit for pagination.
	if err := query.Offset((page - 1) * size).Limit(size).Order("id desc").Find(&permissions).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return permissions, nil
}

// FindWithSort retrieves permissions matching the criteria from the database with sorting support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - sort: sorting criteria (e.g., "id desc").
//
// Returns:
//   - []Permission: slice of retrieved permissions.
//   - error: error if the query fails, otherwise nil.
func (p *Permission) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]Permission, error) {
	var permissions []Permission

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if p.queryCondition != "" {
		query = query.Where(p.queryCondition, p.queryArgs...)
	} else {
		query = query.Where(p)
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&permissions).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return permissions, nil
}
