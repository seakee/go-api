// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package auth provides functionality for authentication and authorization,
// including operations for managing application entities.
package auth

import (
	"context"

	"github.com/seakee/go-api/app/model/auth"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Repo defines the interface for application-related database operations.
type Repo interface {
	// GetApp retrieves an application by its properties.
	GetApp(ctx context.Context, app *auth.App) (*auth.App, error)

	// Create inserts a new application into the database.
	Create(ctx context.Context, app *auth.App) (uint, error)

	// ExistAppByName checks if an application with the given name exists.
	ExistAppByName(ctx context.Context, name string) (bool, error)
}

// repo implements the Repo interface.
type repo struct {
	redis *redis.Manager
	db    *gorm.DB
}

// ExistAppByName checks if an application with the given name exists in the database.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//   - name: The name of the application to check.
//
// Returns:
//   - exist: A boolean indicating whether the application exists.
//   - err: An error if the database operation fails.
//
// Example:
//
//	exists, err := r.ExistAppByName(context.Background(), "MyApp")
//	if err != nil {
//	    log.Printf("Error checking app existence: %v", err)
//	    return
//	}
//	if exists {
//	    fmt.Println("Application already exists")
//	}
func (r repo) ExistAppByName(ctx context.Context, name string) (exist bool, err error) {
	app := &auth.App{AppName: name}
	a, err := app.First(ctx, r.db)
	if a != nil {
		exist = true
	}

	return
}

// Create inserts a new application into the database.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//   - app: A pointer to the auth.App struct containing the application details.
//
// Returns:
//   - uint: The ID of the newly created application.
//   - error: An error if the database operation fails.
//
// Example:
//
//	newApp := &auth.App{AppName: "NewApp", Description: "A new application"}
//	id, err := r.Create(context.Background(), newApp)
//	if err != nil {
//	    log.Printf("Error creating app: %v", err)
//	    return
//	}
//	fmt.Printf("Created new app with ID: %d\n", id)
func (r repo) Create(ctx context.Context, app *auth.App) (uint, error) {
	return app.Create(ctx, r.db)
}

// GetApp retrieves an application from the database based on the provided App struct.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//   - app: A pointer to an auth.App struct with search criteria.
//
// Returns:
//   - *auth.App: A pointer to the retrieved application.
//   - error: An error if the database operation fails.
//
// Example:
//
//	searchApp := &auth.App{ID: 1}
//	foundApp, err := r.GetApp(context.Background(), searchApp)
//	if err != nil {
//	    log.Printf("Error retrieving app: %v", err)
//	    return
//	}
//	fmt.Printf("Found app: %+v\n", foundApp)
func (r repo) GetApp(ctx context.Context, app *auth.App) (*auth.App, error) {
	return app.First(ctx, r.db)
}

// NewAppRepo creates a new instance of the application repository.
//
// Parameters:
//   - db: A pointer to the gorm.DB instance for database operations.
//   - redis: A pointer to the redis.Manager for caching operations.
//
// Returns:
//   - Repo: An implementation of the Repo interface.
//
// Example:
//
//	db := // initialize gorm.DB
//	redisManager := // initialize redis.Manager
//	appRepo := NewAppRepo(db, redisManager)
func NewAppRepo(db *gorm.DB, redis *redis.Manager) Repo {
	return &repo{redis: redis, db: db}
}
