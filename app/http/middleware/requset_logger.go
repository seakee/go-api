// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package middleware provides HTTP middleware functions for the application.
package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

// RequestLogger returns a Gin middleware function that logs details about each HTTP request.
//
// This middleware captures request details such as method, URI, status code, latency,
// client IP, and request body. It logs this information using a structured logger.
//
// Returns:
//   - gin.HandlerFunc: A middleware function for Gin framework.
func (m middleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Read and restore request body
		buf, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

		// Process the request
		c.Next()

		// Record end time and calculate latency
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		// Collect request details
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := util.GetRealIP(c)

		// Get or generate trace ID
		traceID, exists := c.Get("trace_id")
		if !exists {
			traceID = m.traceID.New()
		}

		// Create context with trace ID
		ctx := context.WithValue(context.Background(), logger.TraceIDKey, traceID.(string))

		// Log request details
		m.logger.Info(ctx,
			"Request Logs",
			zap.Int("StatusCode", statusCode),
			zap.Any("Latency", latencyTime),
			zap.String("IP", clientIP),
			zap.String("Method", reqMethod),
			zap.String("RequestPath", reqUri),
			zap.Any("body", string(buf)),
		)
	}
}
