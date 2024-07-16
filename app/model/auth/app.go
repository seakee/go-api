// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type App struct {
	gorm.Model

	AppID       string `gorm:"column:app_id" json:"app_id"`             // Application ID
	AppName     string `gorm:"column:app_name" json:"app_name"`         // Application Name
	AppSecret   string `gorm:"column:app_secret" json:"app_secret"`     // Application Secret Key
	RedirectUri string `gorm:"column:redirect_uri" json:"redirect_uri"` // Redirect URI after authorization
	Description string `gorm:"column:description" json:"description"`   // Description
	Status      int8   `gorm:"column:status" json:"status"`             // 1: Active; 2: Disabled
}

// TableName specifies the table name for the App model.
func (a *App) TableName() string {
	return "auth_app"
}

// First retrieves the first app matching the criteria from the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - *App: pointer to the retrieved app, or nil if not found.
// - error: error if the query fails, otherwise nil.
func (a *App) First(ctx context.Context, db *gorm.DB) (*App, error) {
	var app App

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(a).First(&app).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &app, nil
}

// Last retrieves the last app matching the criteria from the database, ordered by ID in descending order.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - *App: pointer to the retrieved app, or nil if not found.
// - error: error if the query fails, otherwise nil.
func (a *App) Last(ctx context.Context, db *gorm.DB) (*App, error) {
	var app App

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(a).Order("id desc").First(&app).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &app, nil
}

// Create inserts a new app into the database and returns the ID of the created App.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - uint: ID of the created app.
// - error: error if the insert operation fails, otherwise nil.
func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create(a).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return a.ID, nil
}

// Delete removes the app from the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - error: error if the delete operation fails, otherwise nil.
func (a *App) Delete(ctx context.Context, db *gorm.DB) error {
	// Perform the database delete operation with context.
	return db.WithContext(ctx).Delete(a).Error
}

// Updates applies the specified updates to the app in the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - updates: map[string]interface{} containing the updates to apply.
//
// Returns:
// - error: error if the update operation fails, otherwise nil.
func (a *App) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Perform the database update operation with context.
	return db.WithContext(ctx).Model(a).Updates(updates).Error
}

// List retrieves all apps matching the criteria from the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - []App: slice of retrieved apps.
// - error: error if the query fails, otherwise nil.
func (a *App) List(ctx context.Context, db *gorm.DB) ([]App, error) {
	var apps []App

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(a).Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return apps, nil
}

// ListByArgs retrieves apps matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - query: SQL query string.
// - args: variadic arguments for the SQL query.
//
// Returns:
// - []App: slice of retrieved apps.
// - error: error if the query fails, otherwise nil.
func (a *App) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]App, error) {
	var apps []App

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return apps, nil
}

// CountByArgs counts the number of apps matching the specified query and arguments in the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - query: SQL query string.
// - args: variadic arguments for the SQL query.
//
// Returns:
// - int64: count of matching apps.
// - error: error if the count operation fails, otherwise nil.
func (a *App) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&App{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of apps matching the criteria in the database.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
//
// Returns:
// - int64: count of matching apps.
// - error: error if the count operation fails, otherwise nil.
func (a *App) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&App{}).Where(a).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// BatchInsert inserts multiple apps into the database in a single batch operation.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - apps: slice of App instances to be inserted.
//
// Returns:
// - error: error if the batch insert operation fails, otherwise nil.
func (a *App) BatchInsert(ctx context.Context, db *gorm.DB, apps []App) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&apps).Error
}

// FindWithPagination retrieves apps matching the criteria from the database with pagination support.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - page: page number for pagination (1-based).
// - size: number of apps per page.
//
// Returns:
// - []App: slice of retrieved apps.
// - error: error if the query fails, otherwise nil.
func (a *App) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]App, error) {
	var apps []App

	// Perform the database query with context, applying offset and limit for pagination.
	if err := db.WithContext(ctx).Where(a).Offset((page - 1) * size).Limit(size).Find(&apps).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return apps, nil
}

// FindWithSort retrieves apps matching the criteria from the database with sorting support.
//
// Parameters:
// - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// - db: *gorm.DB database connection.
// - sort: sorting criteria (e.g., "id desc").
//
// Returns:
// - []App: slice of retrieved apps.
// - error: error if the query fails, otherwise nil.
func (a *App) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]App, error) {
	var apps []App

	// Perform the database query with context, applying the specified sort order.
	if err := db.WithContext(ctx).Where(a).Order(sort).Find(&apps).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return apps, nil
}
