package middleware

import (
	"github.com/gin-gonic/gin"
)

func (m middleware) SetTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = m.traceID.New()
		}

		c.Set("trace_id", traceID)

		c.Next()
	}
}
