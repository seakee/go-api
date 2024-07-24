// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package job provides functionality for registering and managing scheduled jobs in the application.
// It includes various monitoring and maintenance tasks that run on a regular basis.
package job

import (
	"github.com/seakee/go-api/app/job/monitor"
	"github.com/seakee/go-api/app/pkg/schedule"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Register sets up and schedules all the jobs for the application.
//
// Parameters:
//   - logger: A pointer to the logger.Manager for logging purposes.
//   - redis: A map of Redis managers, keyed by their names.
//   - db: A map of GORM database connections, keyed by their names.
//   - feishu: A pointer to the Feishu manager for Feishu-related operations.
//   - s: A pointer to the schedule.Schedule instance for job scheduling.
//
// This function initializes various monitoring jobs and adds them to the scheduler.
// Currently, it sets up an IP monitor job that runs every 5 minutes without overlapping.
func Register(logger *logger.Manager, redis map[string]*redis.Manager, db map[string]*gorm.DB, feishu *feishu.Manager, s *schedule.Schedule) {
	// Initialize the IP monitor
	ipMonitor := monitor.NewIpMonitor(logger, redis["go-api"])

	// Add the IP monitor job to the scheduler
	// It will run every 5 minutes without overlapping with previous executions
	s.AddJob("IpMonitor", ipMonitor).PerMinuit(5).WithoutOverlapping()
}
