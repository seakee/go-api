// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

// RequestLogger builds a Gin middleware that emits one structured log per request.
//
// Security and observability behavior:
//   - It captures method, URI, status, latency, client IP, and request body.
//   - The request body is read once before handlers run, then restored so downstream
//     handlers can still consume it normally.
//   - Sensitive fields inside supported body formats are masked before logging.
//   - A trace ID is propagated into the logger context so request logs can be
//     correlated across middleware and service layers.
//
// Supported body sanitization:
//   - application/json
//   - application/x-www-form-urlencoded
//
// For unsupported content types, the logger records only an omitted-body marker
// with byte length instead of raw content.
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
		ctx := c.Request.Context()
		if ctx.Value(logger.TraceIDKey) == nil {
			ctx = context.WithValue(ctx, logger.TraceIDKey, traceID.(string))
		}

		body := sanitizeRequestBody(c.GetHeader("Content-Type"), buf)

		// Log request details
		m.logger.Info(ctx,
			"Request Logs",
			zap.Int("StatusCode", statusCode),
			zap.Any("Latency", latencyTime),
			zap.String("IP", clientIP),
			zap.String("Method", reqMethod),
			zap.String("RequestPath", reqUri),
			zap.Any("body", body),
		)
	}
}

// sanitizeRequestBody converts request bytes into a safe-to-log representation.
//
// The function never returns raw secret material for recognized payload formats:
//   - JSON payloads are decoded into generic objects and recursively redacted.
//   - Form payloads are parsed by key/value and sensitive keys are replaced.
//
// If parsing fails or the content type is not supported, the function returns a
// compact marker string that includes only the body size.
func sanitizeRequestBody(contentType string, body []byte) any {
	if len(body) == 0 {
		return ""
	}

	lowerContentType := strings.ToLower(contentType)

	if strings.Contains(lowerContentType, "application/json") {
		var payload any
		if err := json.Unmarshal(body, &payload); err == nil {
			redactSensitiveValue(payload)
			return payload
		}
	}

	if strings.Contains(lowerContentType, "application/x-www-form-urlencoded") {
		values, err := url.ParseQuery(string(body))
		if err == nil {
			result := make(map[string]any, len(values))
			for k, v := range values {
				if isSensitiveField(k) {
					result[k] = "[REDACTED]"
					continue
				}
				result[k] = strings.Join(v, ", ")
			}
			return result
		}
	}

	return fmt.Sprintf("[omitted body, %d bytes]", len(body))
}

// redactSensitiveValue walks nested JSON-like structures and masks secrets in place.
//
// It supports maps and arrays produced by json.Unmarshal on arbitrary payloads.
// Any key identified by isSensitiveField is replaced with "[REDACTED]". Non-container
// scalar values are left unchanged.
func redactSensitiveValue(value any) {
	switch v := value.(type) {
	case map[string]any:
		for key, nested := range v {
			if isSensitiveField(key) {
				v[key] = "[REDACTED]"
				continue
			}
			redactSensitiveValue(nested)
		}
	case []any:
		for i := range v {
			redactSensitiveValue(v[i])
		}
	}
}

// isSensitiveField determines whether a key likely contains credential data.
//
// Matching strategy:
//   - Exact matches for common authentication and secret field names.
//   - Heuristic substring matches to catch custom naming variants.
//
// The input is normalized to lower case and trimmed before evaluation.
func isSensitiveField(field string) bool {
	key := strings.ToLower(strings.TrimSpace(field))
	if key == "" {
		return false
	}

	switch key {
	case "password", "old_password", "new_password", "token", "access_token",
		"refresh_token", "authorization", "app_secret", "secret",
		"credentials", "safe_code", "totp_code", "jwt":
		return true
	}

	return strings.Contains(key, "password") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "credential") ||
		strings.Contains(key, "authorization") ||
		strings.Contains(key, "totp")
}
