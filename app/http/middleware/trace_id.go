// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"github.com/gin-gonic/gin"
)

func (m middleware) SetTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = m.traceID.New()
			c.Writer.Header().Set("X-Trace-ID", traceID)
		}

		c.Set("trace_id", traceID)

		c.Next()
	}
}
