// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package schedule provides functionality for scheduling and managing jobs
// in a distributed environment. It supports various scheduling patterns
// and includes features like job locking and distributed execution.
package schedule

import (
	"time"

	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
)

// Schedule represents the main scheduler structure.
type Schedule struct {
	Logger  *logger.Manager // Logger for the scheduler
	Redis   *redis.Manager  // Redis client for distributed locking
	Job     []*Job          // Slice of jobs managed by this scheduler
	TraceID *trace.ID       // TraceID for logging and tracking
}

// New creates and returns a new Schedule instance.
//
// Parameters:
//   - logger: A logger.Manager instance for logging
//   - redis: A redis.Manager instance for Redis operations
//   - traceID: A trace.ID instance for request tracing
//
// Returns:
//   - *Schedule: A new Schedule instance
//
// Example:
//
//	scheduler := New(loggerInstance, redisInstance, traceIDInstance)
func New(logger *logger.Manager, redis *redis.Manager, traceID *trace.ID) *Schedule {
	return &Schedule{
		Logger:  logger,
		Redis:   redis,
		Job:     make([]*Job, 0),
		TraceID: traceID,
	}
}

// AddJob adds a new job to the scheduler.
//
// Parameters:
//   - name: A string representing the name of the job
//   - handlerFunc: A HandlerFunc that defines the job's execution logic
//
// Returns:
//   - *Job: The newly created Job instance
//
// Example:
//
//	job := scheduler.AddJob("dailyReport", reportHandler)
func (s *Schedule) AddJob(name string, handlerFunc HandlerFunc) *Job {
	return s.addJob(name, handlerFunc)
}

// addJob is an internal method that creates a new Job instance and adds it to the scheduler.
//
// Parameters:
//   - name: A string representing the name of the job
//   - handlerFunc: A HandlerFunc that defines the job's execution logic
//
// Returns:
//   - *Job: The newly created Job instance
func (s *Schedule) addJob(name string, handlerFunc HandlerFunc) *Job {
	// Create a new Job instance with default settings
	j := &Job{
		Name:                  name,
		Logger:                s.Logger,
		Redis:                 s.Redis,
		Handler:               handlerFunc,
		EnableMultipleServers: true,
		EnableOverlapping:     true,
		RunTime:               &RunTime{Done: make(chan struct{})},
		TraceID:               s.TraceID,
	}

	// Add the new job to the scheduler's job slice
	s.Job = append(s.Job, j)

	return j
}

// Start begins the scheduling process for all added jobs.
//
// This method starts a goroutine that ticks every second and
// attempts to run each job in the scheduler.
func (s *Schedule) Start() {
	go func() {
		// Create a ticker that fires every second
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			// Attempt to run each job in the scheduler
			for _, j := range s.Job {
				j.run()
			}
		}
	}()
}
