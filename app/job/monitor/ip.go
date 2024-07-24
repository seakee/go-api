// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package monitor provides monitoring functionalities for various system aspects.
// It includes jobs for monitoring IP changes and other system metrics.
package monitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/seakee/go-api/app/pkg/schedule"
	"github.com/sk-pkg/feishu"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
)

const (
	// CheckCNIpApi is the API endpoint for checking the IP address (Chinese server)
	CheckCNIpApi = "http://members.3322.org/dyndns/getip"
	// CheckIpApi is an alternative API endpoint for checking the IP address
	CheckIpApi = "http://whatismyip.akamai.com/"
	// lastIpKey is the Redis key for storing the last known IP address
	lastIpKey = "monitor:ip:lastIp"
)

// ipHandler is a struct that implements the schedule.HandlerFunc interface
// for monitoring IP address changes.
type ipHandler struct {
	done   chan struct{}
	error  chan error
	logger *logger.Manager
	redis  *redis.Manager
	lastIp string
	feishu *feishu.Manager
}

// setLastIp retrieves the last known IP address from Redis and sets it in the handler.
// If an error occurs, it sends the error to the error channel.
func (ih *ipHandler) setLastIp() {
	lastIp, err := ih.redis.GetString(lastIpKey)
	if err != nil {
		ih.error <- fmt.Errorf("failed to get last IP from Redis: %w", err)
		return
	}

	ih.lastIp = lastIp
}

// Exec is the main execution function for the IP monitor job.
// It checks the current IP address and compares it with the last known IP.
// If a change is detected, it updates the Redis store and logs the change.
//
// Parameters:
//   - ctx: A context.Context for handling cancellation and timeouts.
//
// This function sends a done signal when it completes, regardless of whether an error occurred.
func (ih *ipHandler) Exec(ctx context.Context) {
	// Set the last known IP from Redis
	ih.setLastIp()

	// Create a new HTTP client
	client := resty.New()

	// Make a GET request to check the current IP
	res, err := client.R().Get(CheckCNIpApi)
	if err == nil && res != nil && res.StatusCode() == 200 {
		// Trim any newline characters from the response
		currentIp := strings.TrimRight(string(res.Body()), "\n")

		// Check if the IP has changed
		if ih.lastIp != currentIp && currentIp != "" {
			ih.logger.Info(ctx, "IP has changed", zap.String("last ip", ih.lastIp), zap.String("current ip", currentIp))
			ih.lastIp = currentIp

			// Update the last IP in Redis
			if err = ih.redis.SetString(lastIpKey, currentIp, 0); err != nil {
				ih.error <- fmt.Errorf("failed to set last IP (%s) in Redis: %w", currentIp, err)
			}
		}
	} else if err != nil {
		ih.error <- fmt.Errorf("failed to check IP from %s: %w", CheckCNIpApi, err)
	}

	// Signal that the job is done
	ih.done <- struct{}{}
}

// Error returns the channel for receiving error messages from the handler.
func (ih *ipHandler) Error() <-chan error {
	return ih.error
}

// Done returns the channel for receiving the completion signal from the handler.
func (ih *ipHandler) Done() <-chan struct{} {
	return ih.done
}

// NewIpMonitor creates and returns a new IP monitor handler.
//
// Parameters:
//   - logger: A pointer to the logger.Manager for logging purposes.
//   - redis: A pointer to the redis.Manager for Redis operations.
//
// Returns:
//   - schedule.HandlerFunc: A handler function that can be scheduled for execution.
func NewIpMonitor(logger *logger.Manager, redis *redis.Manager) schedule.HandlerFunc {
	return &ipHandler{
		done:   make(chan struct{}),
		error:  make(chan error),
		logger: logger,
		lastIp: "",
		redis:  redis,
	}
}
