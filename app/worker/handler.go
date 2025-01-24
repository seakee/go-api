// Copyright 2025 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package worker implements background task processing functionality.
// It provides a framework for running background workers that can handle
// various tasks like processing messages, scheduled jobs, and other
// asynchronous operations.
package worker

import (
	"context"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/notify"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Handler defines the interface for worker operations.
// Any type implementing this interface can be used as a worker.
type Handler interface {
	// Start initiates the worker's operations with the given context.
	// Returns an error if the worker fails to start.
	Start(ctx context.Context) error
}

// handler implements the Handler interface and contains all dependencies
// required for worker operations.
type handler struct {
	// logger handles all logging operations
	logger *logger.Manager
	// redis contains named Redis connection managers
	redis map[string]*redis.Manager
	// db contains named database connections
	db map[string]*gorm.DB
	// notify handles notification operations
	notify *notify.Manager
	// traceID handles request tracing
	traceID *trace.ID
}

// Start implements the Handler interface and begins worker operations.
// Currently implementation is a placeholder that returns nil.
// When uncommented, it initializes and runs a Telegram worker.
//
// Parameters:
//   - ctx: context.Context - Context for cancellation and timeout
//
// Returns:
//   - error: Returns nil if successful, error otherwise
func (h *handler) Start(ctx context.Context) error {
	// TODO: Uncomment and implement telegram worker
	// Initialize telegram worker with required dependencies
	// telegramWork, err := telegram.NewWorker(h.logger, h.redis["dudu"], h.db["dudu"], h.notify, h.traceID)
	// if err != nil {
	//     return err
	// }
	//
	// Start telegram worker with provided context
	// telegramWork.Run(ctx)

	return nil
}

// New creates and returns a new Handler instance with the provided dependencies.
//
// Parameters:
//   - logger: *logger.Manager - Logger instance for the worker
//   - redis: map[string]*redis.Manager - Map of named Redis connection managers
//   - db: map[string]*gorm.DB - Map of named database connections
//   - notify: *notify.Manager - Notification manager instance
//   - traceID: *trace.ID - Request tracer instance
//
// Returns:
//   - Handler: New handler instance implementing the Handler interface
func New(logger *logger.Manager, redis map[string]*redis.Manager, db map[string]*gorm.DB, notify *notify.Manager, traceID *trace.ID) Handler {
	// Initialize and return new handler instance with provided dependencies
	return &handler{
		logger:  logger,
		redis:   redis,
		db:      db,
		notify:  notify,
		traceID: traceID,
	}
}
