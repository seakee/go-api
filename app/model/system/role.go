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

// Role represents a role entity in the system RBAC model.
type Role struct {
	gorm.Model

	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`

	// Query condition fields for Where method
	queryCondition string        `gorm:"-" json:"-"`
	queryArgs      []interface{} `gorm:"-" json:"-"`
}

// TableName specifies the table name for the Role model.
func (r *Role) TableName() string {
	return "sys_role"
}

// Where sets query conditions for the Role model.
//
// Parameters:
//   - condition: string SQL condition with placeholders.
//   - args: ...interface{} arguments to replace placeholders in the condition.
//
// Returns:
//   - *Role: pointer to the Role instance for method chaining.
func (r *Role) Where(condition string, args ...interface{}) *Role {
	r.queryCondition = condition
	r.queryArgs = args
	return r
}

// First retrieves the first role matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Role: pointer to the retrieved role, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *Role) First(ctx context.Context, db *gorm.DB) (*Role, error) {
	var role Role

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database query
	if err := query.First(&role).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &role, nil
}

// Last retrieves the last role matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *Role: pointer to the retrieved role, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *Role) Last(ctx context.Context, db *gorm.DB) (*Role, error) {
	var role Role

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database query
	if err := query.Order("id desc").First(&role).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &role, nil
}

// Create inserts a new role into the database and returns the ID of the created Role.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - uint: ID of the created role.
//   - error: error if the insert operation fails, otherwise nil.
func (r *Role) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(r).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return r.ID, nil
}

// Delete removes the role from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - error: error if the delete operation fails, otherwise nil.
func (r *Role) Delete(ctx context.Context, db *gorm.DB) error {
	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		return query.Where(r.queryCondition, r.queryArgs...).Delete(&Role{}).Error
	} else {
		return query.Where(r).Delete(&Role{}).Error
	}
}

// Updates applies the specified updates to the role in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
//   - error: error if the update operation fails, otherwise nil.
func (r *Role) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query with context and explicitly set the model
	query := db.WithContext(ctx).Model(&Role{})

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		return query.Where(r.queryCondition, r.queryArgs...).Updates(updates).Error
	} else {
		return query.Where(r).Updates(updates).Error
	}
}

// List retrieves all roles matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - []Role: slice of retrieved roles.
//   - error: error if the query fails, otherwise nil.
func (r *Role) List(ctx context.Context, db *gorm.DB) ([]Role, error) {
	var roles []Role

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database query
	if err := query.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return roles, nil
}

// ListByArgs retrieves roles matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - []Role: slice of retrieved roles.
//   - error: error if the query fails, otherwise nil.
func (r *Role) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]Role, error) {
	var roles []Role

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return roles, nil
}

// CountByArgs counts the number of roles matching the specified query and arguments in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - int64: count of matching roles.
//   - error: error if the count operation fails, otherwise nil.
func (r *Role) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&Role{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of roles matching the criteria in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: count of matching roles.
//   - error: error if the count operation fails, otherwise nil.
func (r *Role) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Build query with context
	query := db.WithContext(ctx).Model(&Role{})

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database count operation
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// BatchInsert inserts multiple roles into the database in a single batch operation.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - roles: slice of Role instances to be inserted.
//
// Returns:
//   - error: error if the batch insert operation fails, otherwise nil.
func (r *Role) BatchInsert(ctx context.Context, db *gorm.DB, roles []Role) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&roles).Error
}

// FindWithPagination retrieves roles matching the criteria from the database with pagination support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - page: page number for pagination (1-based).
//   - size: number of roles per page.
//
// Returns:
//   - []Role: slice of retrieved roles.
//   - error: error if the query fails, otherwise nil.
func (r *Role) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]Role, error) {
	var roles []Role

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database query with context, applying offset and limit for pagination.
	if err := query.Offset((page - 1) * size).Limit(size).Order("id desc").Find(&roles).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return roles, nil
}

// FindWithSort retrieves roles matching the criteria from the database with sorting support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - sort: sorting criteria (e.g., "id desc").
//
// Returns:
//   - []Role: slice of retrieved roles.
//   - error: error if the query fails, otherwise nil.
func (r *Role) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]Role, error) {
	var roles []Role

	// Build query with context
	query := db.WithContext(ctx)

	// Apply Where condition if set, otherwise use struct fields
	if r.queryCondition != "" {
		query = query.Where(r.queryCondition, r.queryArgs...)
	} else {
		query = query.Where(r)
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&roles).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return roles, nil
}
