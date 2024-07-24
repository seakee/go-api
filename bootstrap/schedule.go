// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

import (
	"context"

	"github.com/seakee/go-api/app/job"
	"github.com/seakee/go-api/app/pkg/schedule"
)

// startSchedule initializes and starts the application's scheduling system.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//
// This function creates a new scheduler, registers jobs, and starts the scheduler.
// It uses the application's logger, Redis connection, and TraceID for creating the scheduler.
func (a *App) startSchedule(ctx context.Context) {
	// Create a new scheduler instance
	s := schedule.New(a.Logger, a.Redis["go-api"], a.TraceID)

	// Register jobs with the scheduler
	// This function call sets up all the scheduled jobs for the application
	job.Register(a.Logger, a.Redis, a.MysqlDB, a.Feishu, s)

	// Start the scheduler
	// This will begin executing the registered jobs according to their schedules
	s.Start()

	// Log successful loading of the scheduler
	a.Logger.Info(ctx, "Schedule loaded successfully")
}
