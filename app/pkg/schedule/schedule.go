// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package schedule

import (
	"time"

	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
)

type (
	Schedule struct {
		Logger  *logger.Manager
		Redis   *redis.Manager
		Job     []*Job
		TraceID *trace.ID
	}
)

func New(logger *logger.Manager, redis *redis.Manager, traceID *trace.ID) *Schedule {
	return &Schedule{
		Logger:  logger,
		Redis:   redis,
		Job:     make([]*Job, 0),
		TraceID: traceID,
	}
}

func (s *Schedule) AddJob(name string, handlerFunc HandlerFunc) *Job {
	return s.addJob(name, handlerFunc)
}

func (s *Schedule) addJob(name string, handlerFunc HandlerFunc) *Job {
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

	s.Job = append(s.Job, j)

	return j
}

func (s *Schedule) Start() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			for _, j := range s.Job {
				j.run()
			}
		}
	}()
}
