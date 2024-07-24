// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"github.com/gin-gonic/gin"
)

// SetTraceID returns a Gin middleware function that sets a trace ID for each request.
//
// This middleware checks for an existing trace ID in the "X-Trace-ID" header.
// If not present, it generates a new trace ID. The trace ID is then set in both
// the response header and the Gin context.
//
// Returns:
//   - gin.HandlerFunc: A middleware function for Gin framework.
func (m middleware) SetTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for existing trace ID in header
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			// Generate new trace ID if not present
			traceID = m.traceID.New()
			c.Writer.Header().Set("X-Trace-ID", traceID)
		}

		// Set trace ID in Gin context
		c.Set("trace_id", traceID)

		// Continue to the next middleware or handler
		c.Next()
	}
}
