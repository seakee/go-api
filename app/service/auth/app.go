// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/seakee/go-api/app/model/auth"
	authRepo "github.com/seakee/go-api/app/repository/auth"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"gorm.io/gorm"
)

// AppService defines the interface for app-related business operations.
type AppService interface {
	// CreateApp creates a new app with the provided parameters.
	CreateApp(ctx context.Context, params *CreateAppParams) (*CreateAppResult, error)

	// GetTokenByCredentials validates app credentials and returns token information.
	GetTokenByCredentials(ctx context.Context, appID, appSecret string) (*auth.App, error)

	// GetAppByID retrieves an app by its ID.
	GetAppByID(ctx context.Context, id uint) (*auth.App, error)

	// UpdateApp updates an existing app.
	UpdateApp(ctx context.Context, id uint, params *UpdateAppParams) error

	// DeleteApp deletes an app by its ID.
	DeleteApp(ctx context.Context, id uint) error

	// ListApps retrieves a list of apps based on query conditions.
	ListApps(ctx context.Context, params *ListAppParams) ([]auth.App, error)
}

// CreateAppParams defines the parameters for creating a new app.
type CreateAppParams struct {
	AppName     string `json:"app_name" binding:"required"`
	Description string `json:"description"`
	RedirectUri string `json:"redirect_uri"`
}

// UpdateAppParams defines the parameters for updating an app.
type UpdateAppParams struct {
	AppName     string `json:"app_name"`
	Description string `json:"description"`
	RedirectUri string `json:"redirect_uri"`
	Status      int8   `json:"status"`
}

// ListAppParams defines the parameters for listing apps.
type ListAppParams struct {
	AppName string `json:"app_name"`
	Status  int8   `json:"status"`
}

// CreateAppResult defines the result returned when creating an app.
type CreateAppResult struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// appService implements the AppService interface.
type appService struct {
	repo authRepo.AppRepo
}

// NewAppService creates a new instance of the app service.
//
// Parameters:
//   - db: *gorm.DB database connection.
//   - redis: *redis.Manager redis connection.
//
// Returns:
//   - AppService: An implementation of the AppService interface.
func NewAppService(db *gorm.DB, redis *redis.Manager) AppService {
	// Initialize repository with database and redis connections
	repo := authRepo.NewAppRepo(db, redis)

	return &appService{
		repo: repo,
	}
}

// CreateApp creates a new app with the provided parameters.
//
// This method handles the business logic for creating a new app:
// 1. Validates that an app with the given name doesn't already exist.
// 2. Generates a unique app ID and secret.
// 3. Creates the app record in the database.
// 4. Returns the generated credentials.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - params: *CreateAppParams the parameters for creating the app.
//
// Returns:
//   - *CreateAppResult: the result containing app ID and secret.
//   - error: error if the creation fails, otherwise nil.
func (s *appService) CreateApp(ctx context.Context, params *CreateAppParams) (*CreateAppResult, error) {
	// Check if app already exists
	exists, err := s.repo.ExistAppByName(ctx, params.AppName)
	if err != nil {
		return nil, fmt.Errorf("failed to check app existence: %w", err)
	}

	if exists {
		return nil, errors.New("app with this name already exists")
	}

	// Create new app
	app := &auth.App{
		AppName:     params.AppName,
		AppId:       "go-api-" + util.RandLowStr(8),
		AppSecret:   util.RandUpStr(32),
		RedirectUri: params.RedirectUri,
		Description: params.Description,
		Status:      1, // Active by default
	}

	// Save app to repository
	_, err = s.repo.Create(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	// Return credentials
	return &CreateAppResult{
		AppID:     app.AppId,
		AppSecret: app.AppSecret,
	}, nil
}

// GetTokenByCredentials validates app credentials and returns the app information.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - appID: string the app ID.
//   - appSecret: string the app secret.
//
// Returns:
//   - *auth.App: the app information if credentials are valid.
//   - error: error if validation fails, otherwise nil.
func (s *appService) GetTokenByCredentials(ctx context.Context, appID, appSecret string) (*auth.App, error) {
	if appID == "" || appSecret == "" {
		return nil, errors.New("app_id and app_secret are required")
	}

	app, err := s.repo.GetAppByCredentials(ctx, appID, appSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to validate credentials: %w", err)
	}

	if app == nil {
		return nil, errors.New("invalid credentials")
	}

	if app.Status != 1 {
		return nil, errors.New("app is not active")
	}

	return app, nil
}

// GetAppByID retrieves an app by its ID.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app.
//
// Returns:
//   - *auth.App: the app information.
//   - error: error if retrieval fails, otherwise nil.
func (s *appService) GetAppByID(ctx context.Context, id uint) (*auth.App, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get app by ID: %w", err)
	}

	if app == nil {
		return nil, errors.New("app not found")
	}

	return app, nil
}

// UpdateApp updates an existing app.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app to update.
//   - params: *UpdateAppParams the update parameters.
//
// Returns:
//   - error: error if update fails, otherwise nil.
func (s *appService) UpdateApp(ctx context.Context, id uint, params *UpdateAppParams) error {
	// Check if app exists
	existingApp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	if existingApp == nil {
		return errors.New("app not found")
	}

	// If changing app name, check for duplicates
	if params.AppName != "" && params.AppName != existingApp.AppName {
		exists, err := s.repo.ExistAppByName(ctx, params.AppName)
		if err != nil {
			return fmt.Errorf("failed to check app name existence: %w", err)
		}
		if exists {
			return errors.New("app with this name already exists")
		}
	}

	// Update app
	updateApp := &auth.App{
		AppName:     params.AppName,
		Description: params.Description,
		RedirectUri: params.RedirectUri,
		Status:      params.Status,
	}

	err = s.repo.Update(ctx, id, updateApp)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}

	return nil
}

// DeleteApp deletes an app by its ID.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the app to delete.
//
// Returns:
//   - error: error if deletion fails, otherwise nil.
func (s *appService) DeleteApp(ctx context.Context, id uint) error {
	// Check if app exists
	existingApp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	if existingApp == nil {
		return errors.New("app not found")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	return nil
}

// ListApps retrieves a list of apps based on query conditions.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - params: *ListAppParams the query parameters.
//
// Returns:
//   - []auth.App: slice of apps.
//   - error: error if query fails, otherwise nil.
func (s *appService) ListApps(ctx context.Context, params *ListAppParams) ([]auth.App, error) {
	queryApp := &auth.App{}

	if params.AppName != "" {
		queryApp.AppName = params.AppName
	}

	if params.Status != 0 {
		queryApp.Status = params.Status
	}

	apps, err := s.repo.List(ctx, queryApp)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	return apps, nil
}
