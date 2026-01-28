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

// User represents a user account in the system.
type User struct {
	gorm.Model

	Account     string `gorm:"column:account" json:"account"`
	Password    string `gorm:"column:password" json:"password"`
	Salt        string `gorm:"column:salt" json:"salt"`
	TotpKey     string `gorm:"column:totp_key" json:"totp_key"`
	TotpEnabled bool   `gorm:"column:totp_enabled" json:"totp_enabled"`
	FeishuId    string `gorm:"column:feishu_id" json:"feishu_id"`
	WechatId    string `gorm:"column:wechat_id" json:"wechat_id"`
	GithubId    string `gorm:"column:github_id" json:"github_id"`
	UserName    string `gorm:"column:user_name" json:"user_name"`
	Status      int8   `gorm:"column:status" json:"status"`
	Avatar      string `gorm:"column:avatar" json:"avatar"`

	// Fields for Where support
	queryCondition string
	queryArgs      []interface{}
}

// Where sets query condition and arguments for the User instance.
//
// Parameters:
//   - query: SQL query string or condition.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - *User: pointer to the User instance for method chaining.
func (u *User) Where(query string, args ...interface{}) *User {
	u.queryCondition = query
	u.queryArgs = args
	return u
}

// TableName specifies the table name for the User model.
func (u *User) TableName() string {
	return "sys_user"
}

// First retrieves the first user matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *User: pointer to the retrieved user, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (u *User) First(ctx context.Context, db *gorm.DB) (*User, error) {
	var user User

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.First(&user).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &user, nil
}

// Last retrieves the last user matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - *User: pointer to the retrieved user, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (u *User) Last(ctx context.Context, db *gorm.DB) (*User, error) {
	var user User

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.Order("id desc").First(&user).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &user, nil
}

// Create inserts a new user into the database and returns the ID of the created User.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - uint: ID of the created user.
//   - error: error if the insert operation fails, otherwise nil.
func (u *User) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return u.ID, nil
}

// Delete removes the user from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - error: error if the delete operation fails, otherwise nil.
func (u *User) Delete(ctx context.Context, db *gorm.DB) error {
	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database delete operation with context.
	return query.Delete(u).Error
}

// Updates applies the specified updates to the user in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
//   - error: error if the update operation fails, otherwise nil.
func (u *User) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query based on whether Where was used
	query := db.WithContext(ctx).Model(u)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database update operation with context.
	return query.Updates(updates).Error
}

// List retrieves all users matching the criteria from the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - []User: slice of retrieved users.
//   - error: error if the query fails, otherwise nil.
func (u *User) List(ctx context.Context, db *gorm.DB) ([]User, error) {
	var users []User

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context.
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return users, nil
}

// ListByArgs retrieves users matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - []User: slice of retrieved users.
//   - error: error if the query fails, otherwise nil.
func (u *User) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]User, error) {
	var users []User

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return users, nil
}

// CountByArgs counts the number of users matching the specified query and arguments in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - query: SQL query string.
//   - args: variadic arguments for the SQL query.
//
// Returns:
//   - int64: count of matching users.
//   - error: error if the count operation fails, otherwise nil.
func (u *User) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&User{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of users matching the criteria in the database.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//
// Returns:
//   - int64: count of matching users.
//   - error: error if the count operation fails, otherwise nil.
func (u *User) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Build query based on whether Where was used
	query := db.WithContext(ctx).Model(&User{})
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

// BatchInsert inserts multiple users into the database in a single batch operation.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - users: slice of User instances to be inserted.
//
// Returns:
//   - error: error if the batch insert operation fails, otherwise nil.
func (u *User) BatchInsert(ctx context.Context, db *gorm.DB, users []User) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&users).Error
}

// FindWithPagination retrieves users matching the criteria from the database with pagination support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - page: page number for pagination (1-based).
//   - size: number of users per page.
//
// Returns:
//   - []User: slice of retrieved users.
//   - error: error if the query fails, otherwise nil.
func (u *User) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]User, error) {
	var users []User

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context, applying offset and limit for pagination.
	if err := query.Offset((page - 1) * size).Limit(size).Order("id desc").Find(&users).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return users, nil
}

// FindWithSort retrieves users matching the criteria from the database with sorting support.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - db: *gorm.DB database connection.
//   - sort: sorting criteria (e.g., "id desc").
//
// Returns:
//   - []User: slice of retrieved users.
//   - error: error if the query fails, otherwise nil.
func (u *User) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]User, error) {
	var users []User

	// Build query based on whether Where was used
	query := db.WithContext(ctx)
	if u.queryCondition != "" {
		query = query.Where(u.queryCondition, u.queryArgs...)
	} else {
		query = query.Where(u)
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&users).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return users, nil
}
