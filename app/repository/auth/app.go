// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"github.com/seakee/go-api/app/model/auth"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// AppRepo defines the interface for app-related database operations.
type AppRepo interface {
	// GetApp retrieves a app by its properties.
	GetApp(ctx context.Context, app *auth.App) (*auth.App, error)

	// Create inserts a new app into the database.
	Create(ctx context.Context, app *auth.App) (uint, error)

	// Update updates an existing app in the database.
	Update(ctx context.Context, id uint, app *auth.App) error

	// Delete deletes a app by its ID.
	Delete(ctx context.Context, id uint) error

	// List retrieves app records based on query conditions.
	List(ctx context.Context, app *auth.App) ([]auth.App, error)

	// GetByID retrieves a app by its ID.
	GetByID(ctx context.Context, id uint) (*auth.App, error)
}

// appRepo implements the AppRepo interface.
type appRepo struct {
	redis *redis.Manager
	db    *gorm.DB
}

// NewAppRepo creates a new instance of the app repository.
//
// Parameters:
//   - db: A pointer to the gorm.DB instance for database operations.
//   - redis: A pointer to the redis.Manager for caching operations.
//
// Returns:
//   - AppRepo: An implementation of the AppRepo interface.
//
// Example:
//
//	db := // initialize gorm.DB
//	redisManager := // initialize redis.Manager
//	appRepo := NewAppRepo(db, redisManager)
func NewAppRepo(db *gorm.DB, redis *redis.Manager) AppRepo {
	return &appRepo{redis: redis, db: db}
}

// GetApp retrieves a app by its properties using the model's First method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - app: *auth.App the app with search criteria.
//
// Returns:
//   - *auth.App: pointer to the retrieved app, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *appRepo) GetApp(ctx context.Context, app *auth.App) (*auth.App, error) {
	return app.First(ctx, r.db)
}

// Create creates a new app record using the model's Create method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - app: *auth.App the app to create.
//
// Returns:
//   - uint: the ID of the created app.
//   - error: error if the creation fails, otherwise nil.
func (r *appRepo) Create(ctx context.Context, app *auth.App) (uint, error) {
	return app.Create(ctx, r.db)
}

// GetByID retrieves a app by its ID using the model's Where and First methods.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app to retrieve.
//
// Returns:
//   - *auth.App: pointer to the retrieved app, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *appRepo) GetByID(ctx context.Context, id uint) (*auth.App, error) {
	app := &auth.App{}
	return app.Where("id = ?", id).First(ctx, r.db)
}

// Update updates an existing app record using the model's Updates method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app to update.
//   - app: *auth.App the app with updated fields.
//
// Returns:
//   - error: error if the update fails, otherwise nil.
//
// Note: This method will only update non-zero value fields. You need to manually check
// each field and add it to the data map if it's not a zero value.
func (r *appRepo) Update(ctx context.Context, id uint, app *auth.App) error {
	data := make(map[string]interface{})

	if app.AppId != "" {
		data["app_id"] = app.AppId
	}

	if app.AppName != "" {
		data["app_name"] = app.AppName
	}

	if app.AppSecret != "" {
		data["app_secret"] = app.AppSecret
	}

	if app.RedirectUri != "" {
		data["redirect_uri"] = app.RedirectUri
	}

	if app.Description != "" {
		data["description"] = app.Description
	}

	if app.Status != 0 {
		data["status"] = app.Status
	}

	if app.Default != nil {
		data["default"] = app.Default
	}

	if app.Collate != nil {
		data["collate"] = app.Collate
	}

	if len(data) == 0 {
		return nil // No fields to update
	}

	updateModel := &auth.App{}
	updateModel.ID = id

	return updateModel.Updates(ctx, r.db, data)
}

// Delete deletes a app record using the model's Where and Delete methods.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app to delete.
//
// Returns:
//   - error: error if the deletion fails, otherwise nil.
func (r *appRepo) Delete(ctx context.Context, id uint) error {
	app := &auth.App{}
	return app.Where("id = ?", id).Delete(ctx, r.db)
}

// List retrieves app records based on query conditions using the model's List method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - app: *auth.App the app with query conditions.
//
// Returns:
//   - []auth.App: slice of app records.
//   - error: error if the query fails, otherwise nil.
func (r *appRepo) List(ctx context.Context, app *auth.App) ([]auth.App, error) {
	return app.List(ctx, r.db)
}
