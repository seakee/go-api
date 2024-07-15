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

type (
	Repo interface {
		GetApp(ctx context.Context, app *auth.App) (*auth.App, error)
		Create(ctx context.Context, app *auth.App) (uint, error)
		ExistAppByName(ctx context.Context, name string) (bool, error)
	}

	repo struct {
		redis *redis.Manager
		db    *gorm.DB
	}
)

func (r repo) ExistAppByName(ctx context.Context, name string) (exist bool, err error) {
	app := &auth.App{AppName: name}
	a, err := app.First(ctx, r.db)
	if a != nil {
		exist = true
	}

	return
}

func (r repo) Create(ctx context.Context, app *auth.App) (uint, error) {
	return app.Create(ctx, r.db)
}

func (r repo) GetApp(ctx context.Context, app *auth.App) (*auth.App, error) {
	return app.First(ctx, r.db)
}

func NewAppRepo(db *gorm.DB, redis *redis.Manager) Repo {
	return &repo{redis: redis, db: db}
}
