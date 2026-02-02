// Copyright 2025 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"
	"github.com/seakee/go-api/app/worker"
	"go.uber.org/zap"
)

// startWorker initializes and starts the background worker process.
//
// This function creates a new worker instance with the application's dependencies
// and starts it with the provided context. If the worker fails to start,
// it logs a fatal error and terminates the application.
//
// Parameters:
//   - ctx: context.Context - The context for worker execution and cancellation
func (a *App) startWorker(ctx context.Context) {
	// Initialize new worker instance with application dependencies
	h := worker.New(a.Logger, a.Redis, a.SqlDB, a.Notify, a.TraceID)

	// Start the worker and capture any startup errors
	err := h.Start(ctx)
	if err != nil {
		// Log fatal error and terminate if worker fails to start
		a.Logger.Fatal(ctx, "Worker start failed", zap.Error(err))
		return
	}

	// Log successful worker startup
	a.Logger.Info(ctx, "Worker loaded successfully")
}
