// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package schedule

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

// RunType represents the type of job execution.
type RunType string

// Constants for job run types and default server lock TTL
const (
	DailyRunType     RunType = "daily"     // Run task daily
	SecondlyRunType  RunType = "seconds"   // Run task every X seconds
	MinutelyRunType  RunType = "minute"    // Run task every X minutes
	HourlyRunType    RunType = "hour"      // Run task every X hours
	ImmediateRunType RunType = "immediate" // Run task immediately

	DefaultServerLockTTL = 600 // Default single-server task lock time (10 minutes)
)

// Job represents a scheduled task with its properties and execution settings.
type Job struct {
	Name                  string          // Name of the job instance
	Logger                *logger.Manager // Logger for the job
	Redis                 *redis.Manager  // Redis client for job operations
	Handler               HandlerFunc     // Function to be executed
	EnableMultipleServers bool            // Allow execution on multiple nodes
	EnableOverlapping     bool            // Allow job to run even if previous instance is still running
	RunTime               *RunTime        // Runtime parameters for the job
	TraceID               *trace.ID       // TraceID for job execution tracking
}

// HandlerFunc interface defines the methods that a job handler must implement.
type HandlerFunc interface {
	Exec(ctx context.Context)
	Error() <-chan error
	Done() <-chan struct{}
}

// RunTime contains the runtime parameters for a job.
type RunTime struct {
	Type          RunType       // Type of schedule (daily, per second, per minute, per hour, immediate)
	Time          interface{}   // Execution time or interval
	Locked        bool          // Execution lock for non-overlapping jobs
	PerTypeLocked bool          // Lock for interval-based job types
	Done          chan struct{} // Channel to signal job completion
	RandomDelay   *RandomDelay  // Random delay settings for job execution
}

// RandomDelay defines the minimum and maximum random delay for job execution.
type RandomDelay struct {
	Min int // Minimum delay in seconds
	Max int // Maximum delay in seconds
}

// WithoutOverlapping sets the job to not allow overlapping executions.
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.WithoutOverlapping()
func (j *Job) WithoutOverlapping() *Job {
	j.EnableOverlapping = false
	return j
}

// RandomDelay sets a random delay range for job execution.
//
// Parameters:
//   - min: Minimum delay in seconds
//   - max: Maximum delay in seconds
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.RandomDelay(30, 60)
func (j *Job) RandomDelay(min, max int) *Job {
	if max < min {
		panic("must max > min")
	}

	j.RunTime.RandomDelay = &RandomDelay{
		Min: min,
		Max: max,
	}

	return j
}

func (j *Job) Immediate() *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = ImmediateRunType
	}

	j.EnableOverlapping = false
	return j
}

// DailyAt schedules the job to run at specific times each day.
//
// Parameters:
//   - time: One or more time strings in "HH:MM:SS" format
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.DailyAt("07:30:00", "12:00:00", "18:00:00")
func (j *Job) DailyAt(time ...string) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = DailyRunType
		j.RunTime.Time = time
	}
	return j
}

// PerSeconds schedules the job to run every specified number of seconds.
//
// Parameters:
//   - seconds: Interval in seconds
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.PerSeconds(30)
func (j *Job) PerSeconds(seconds int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = SecondlyRunType
		j.RunTime.Time = time.Duration(seconds) * time.Second
	}
	return j
}

// PerMinuit schedules the job to run every specified number of minutes.
//
// Parameters:
//   - minuit: Interval in minutes
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.PerMinuit(15)
func (j *Job) PerMinuit(minuit int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = MinutelyRunType
		j.RunTime.Time = time.Duration(minuit) * time.Minute
	}
	return j
}

// PerHour schedules the job to run every specified number of hours.
//
// Parameters:
//   - hour: Interval in hours
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.PerHour(4)
func (j *Job) PerHour(hour int) *Job {
	if j.RunTime.Type == "" {
		j.RunTime.Type = HourlyRunType
		j.RunTime.Time = time.Duration(hour) * time.Hour
	}
	return j
}

// OnOneServer sets the job to run on only one server in a distributed environment.
//
// Returns:
//   - *Job: The modified Job instance
//
// Example:
//
//	job.OnOneServer()
func (j *Job) OnOneServer() *Job {
	j.EnableMultipleServers = false
	return j
}

// runWithRecover executes the job handler with panic recovery.
func (j *Job) runWithRecover() {
	ctx := context.WithValue(context.Background(), logger.TraceIDKey, j.TraceID.New())

	defer func() {
		// Recover from panic and log the error
		if r := recover(); r != nil {
			j.Logger.Error(ctx, "job has a panic error",
				zap.String("error", fmt.Sprintf("%v", r)),
				zap.String("stack", string(debug.Stack())),
			)
		}
	}()

	j.handler(ctx)
}

// run executes the job based on its schedule type.
func (j *Job) run() {
	switch j.RunTime.Type {
	case ImmediateRunType:
		// Run the job immediately
		go j.runWithRecover()
	case DailyRunType:
		// Check if current time matches any of the scheduled times
		times := j.RunTime.Time.([]string)
		for _, t := range times {
			if time.Now().Format("15:04:05") == t {
				go j.runWithRecover()
			}
		}
	case SecondlyRunType, MinutelyRunType, HourlyRunType:
		// Ensure the job is only started once
		if j.RunTime.PerTypeLocked {
			return
		}

		j.RunTime.PerTypeLocked = true
		go func() {
			ticker := time.NewTicker(j.RunTime.Time.(time.Duration))
			for range ticker.C {
				go j.runWithRecover()
			}
		}()
	}
}

// handler manages the job execution process, including locking and error handling.
//
// Parameters:
//   - ctx: Context for the job execution
func (j *Job) handler(ctx context.Context) {
	if !j.EnableOverlapping {
		// Prevent overlapping executions
		if j.RunTime.Locked {
			return
		}
		j.RunTime.Locked = true
	}

	if !j.EnableMultipleServers {
		// Ensure the job runs on only one server
		if !j.lock("Server", DefaultServerLockTTL, false) {
			j.RunTime.Locked = false
			return
		}

		go j.renewalServerLock(ctx)
	}

	// Apply random delay if set
	j.randomDelay()

	j.Logger.Info(ctx, util.SpliceStr("The scheduled job: ", j.Name, " starts execution."))

	// Handle job execution and potential errors
	go func(ctx context.Context) {
	Exit:
		for {
			select {
			case err := <-j.Handler.Error():
				if err != nil {
					j.Logger.Error(ctx, fmt.Sprintf("An error occurred while executing the %s.", j.Name), zap.Error(err))
				}
			case <-j.Handler.Done():
				// Clean up after job completion
				if !j.EnableMultipleServers {
					j.RunTime.Done <- struct{}{}
				}

				j.RunTime.Locked = false

				j.Logger.Info(ctx, util.SpliceStr("The scheduled job: ", j.Name, " has done."))

				break Exit
			}
		}
	}(ctx)

	j.Handler.Exec(ctx)
}

// randomDelay applies a random delay within the specified range before job execution.
func (j *Job) randomDelay() {
	if j.RunTime.RandomDelay == nil {
		return
	}

	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	delay := generator.Intn(j.RunTime.RandomDelay.Max) + j.RunTime.RandomDelay.Min
	time.Sleep(time.Duration(delay) * time.Second)
}

// lock attempts to acquire or renew a Redis lock for the job.
//
// Parameters:
//   - name: Name of the lock
//   - ttl: Time-to-live for the lock in seconds
//   - renewal: Whether this is a lock renewal operation
//
// Returns:
//   - bool: True if the lock was acquired or renewed successfully, false otherwise
func (j *Job) lock(name string, ttl int, renewal bool) bool {
	prefix := j.Redis.Prefix
	key := util.SpliceStr(prefix, "schedule:jobLock:", j.Name, ":", name)

	if renewal {
		_, err := j.Redis.Do("EXPIRE", key, ttl)
		if err == nil {
			return true
		}
	} else {
		ok, err := j.Redis.Do("SET", key, "locked", "EX", ttl, "NX")
		if ok != nil && err == nil {
			return true
		}
	}

	return false
}

// unLock releases the Redis lock for the job.
//
// Parameters:
//   - ctx: Context for logging
//   - name: Name of the lock to release
func (j *Job) unLock(ctx context.Context, name string) {
	key := util.SpliceStr("schedule:jobLock:", j.Name, ":", name)

	ok, err := j.Redis.Del(key)
	if !ok && err != nil {
		j.Logger.Error(ctx, util.SpliceStr("unLock job:", name, "failed"), zap.Error(err))
	}
}

// renewalServerLock periodically renews the server lock to prevent expiration.
//
// Parameters:
//   - ctx: Context for logging and cancellation
func (j *Job) renewalServerLock(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
Exit:
	for {
		select {
		case <-ticker.C:
			// Renew the lock every second
			j.lock("Server", DefaultServerLockTTL, true)
		case <-j.RunTime.Done:
			// Release the lock when the job is done
			j.unLock(ctx, "Server")
			break Exit
		}
	}
}
