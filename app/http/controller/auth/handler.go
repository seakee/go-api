// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package auth provides authentication-related functionality for the application.
package auth

import (
	"context"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/repository/auth"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
)

// Handler interface defines the methods that should be implemented by the auth handler.
type Handler interface {
	i()
	ctx(c *gin.Context) context.Context
	Create() gin.HandlerFunc
	GetToken() gin.HandlerFunc
}

// handler struct implements the Handler interface.
type handler struct {
	logger *logger.Manager
	redis  *redis.Manager
	i18n   *i18n.Manager
	repo   auth.Repo
}

// ctx creates a new context with the trace ID from the gin.Context.
//
// Parameters:
//   - c: *gin.Context - The gin context containing the trace ID.
//
// Returns:
//   - context.Context: A new context with the trace ID added.
func (h handler) ctx(c *gin.Context) context.Context {
	traceID, _ := c.Get("trace_id")

	return context.WithValue(context.Background(), logger.TraceIDKey, traceID.(string))
}

// i is a dummy method to satisfy the Handler interface.
func (h handler) i() {}

// New creates and returns a new Handler instance.
//
// Parameters:
//   - logger: *logger.Manager - The logger manager.
//   - redis: *redis.Manager - The redis manager.
//   - i18n: *i18n.Manager - The i18n manager.
//   - db: *gorm.DB - The database connection.
//
// Returns:
//   - Handler: A new Handler instance.
func New(logger *logger.Manager, redis *redis.Manager, i18n *i18n.Manager, db *gorm.DB) Handler {
	return &handler{
		logger: logger,
		redis:  redis,
		i18n:   i18n,
		repo:   auth.NewAppRepo(db, redis),
	}
}
