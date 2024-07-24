// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package middleware provides HTTP middleware functions for the application.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Middleware interface defines the methods that should be implemented by middleware handlers.
type Middleware interface {
	CheckAppAuth() gin.HandlerFunc
	Cors() gin.HandlerFunc
	RequestLogger() gin.HandlerFunc
	SetTraceID() gin.HandlerFunc
}

// middleware struct implements the Middleware interface.
type middleware struct {
	logger  *logger.Manager
	i18n    *i18n.Manager
	db      map[string]*gorm.DB
	redis   map[string]*redis.Manager
	traceID *trace.ID
}

// New creates and returns a new Middleware instance.
//
// Parameters:
//   - logger: *logger.Manager - The logger manager.
//   - i18n: *i18n.Manager - The internationalization manager.
//   - db: map[string]*gorm.DB - A map of database connections.
//   - redis: map[string]*redis.Manager - A map of Redis managers.
//   - traceID: *trace.ID - The trace ID generator.
//
// Returns:
//   - Middleware: A new Middleware instance.
func New(logger *logger.Manager, i18n *i18n.Manager, db map[string]*gorm.DB, redis map[string]*redis.Manager, traceID *trace.ID) Middleware {
	return &middleware{logger: logger, i18n: i18n, db: db, redis: redis, traceID: traceID}
}
