package schedule

import (
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"time"
)

type (
	Schedule struct {
		Logger *zap.Logger
		Redis  *redis.Manager
		Job    []*Job
	}
)

func New(logger *zap.Logger, redis *redis.Manager) *Schedule {
	return &Schedule{
		Logger: logger,
		Redis:  redis,
		Job:    make([]*Job, 0),
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
