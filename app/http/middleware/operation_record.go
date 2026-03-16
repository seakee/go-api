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
	jsoniter "github.com/json-iterator/go"
	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
)

// SaveOperationRecord creates a middleware for saving operation records.
//
// Returns:
//   - gin.HandlerFunc: Gin middleware function
//
// Example:
//
//	router.Use(middleware.SaveOperationRecord())
func (m middleware) SaveOperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		traceID := m.getOrCreateTraceID(c)
		ctx := context.WithValue(c.Request.Context(), logger.TraceIDKey, traceID)

		params := m.getRequestParams(c)
		writer := newResponseBodyWriter(c.Writer, m.logger)
		c.Writer = writer

		c.Next()

		record := m.createOperationRecord(c, params, traceID, startTime, writer)
		if err := m.operationRecordService.Create(ctx, record); err != nil {
			m.logger.Error(ctx, "create operation record:", zap.Error(err))
		}
	}
}

// getOrCreateTraceID retrieves or creates a trace ID.
func (m middleware) getOrCreateTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return m.traceID.New()
	}
	return traceID.(string)
}

// getRequestParams retrieves request parameters based on request type.
func (m middleware) getRequestParams(c *gin.Context) string {
	if shouldOmitOperationPayload(c.Request.URL.Path) {
		return `{"msg":"request payload omitted for sensitive endpoint"}`
	}

	switch {
	case strings.Contains(c.GetHeader("Content-Type"), "boundary"):
		return `{"msg":"upload file"}`
	case c.Request.ContentLength > 65535:
		return `{"msg":"request body too large"}`
	case c.Request.Method == "GET":
		return sanitizeQueryParams(c.Request.URL.RawQuery)
	default:
		return m.getRawBody(c)
	}
}

// getRawBody retrieves the request body content.
func (m middleware) getRawBody(c *gin.Context) string {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		m.logger.Error(context.Background(), "read body from request:", zap.Error(err))
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	return formatSanitizedPayload(sanitizeRequestBody(c.GetHeader("Content-Type"), body))
}

// createOperationRecord creates an operation record.
func (m middleware) createOperationRecord(c *gin.Context, params, traceID string, startTime time.Time, writer *responseBodyWriter) *system.OperationRecord {
	responseBody := ""
	if c.Request.Method != "GET" {
		if shouldOmitOperationPayload(c.Request.URL.Path) {
			responseBody = `{"msg":"response payload omitted for sensitive endpoint"}`
		} else {
			responseBody = sanitizeResponseBody(writer.body.Bytes())
		}
	}

	record := &system.OperationRecord{
		IP:           util.GetRealIP(c),
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Agent:        c.Request.UserAgent(),
		Params:       params,
		Status:       c.Writer.Status(),
		Latency:      time.Since(startTime).Seconds(),
		Resp:         responseBody,
		ErrorMessage: writer.errMessage,
		TraceID:      traceID,
	}
	if writer.hasStatusCode {
		record.Status = writer.statusCode
	}
	record.CreatedAt = startTime

	if userID, exists := c.Get("user_id"); exists {
		record.UserID = userID.(uint)
	}

	return record
}

// responseBodyWriter is a custom response body writer.
type responseBodyWriter struct {
	gin.ResponseWriter
	body          *bytes.Buffer
	logger        *logger.Manager
	errMessage    string
	statusCode    int
	hasStatusCode bool
}

// newResponseBodyWriter creates a new response body writer.
func newResponseBodyWriter(writer gin.ResponseWriter, logger *logger.Manager) *responseBodyWriter {
	return &responseBodyWriter{
		ResponseWriter: writer,
		body:           &bytes.Buffer{},
		logger:         logger,
	}
}

// Write implements the io.Writer interface for writing response body and recording error messages.
func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	if codeVal := jsoniter.Get(b, "code"); codeVal.ValueType() == jsoniter.NumberValue {
		r.statusCode = codeVal.ToInt()
		r.hasStatusCode = true
	}
	if r.hasStatusCode && r.statusCode != 0 {
		r.errMessage = jsoniter.Get(b, "msg").ToString()
	}
	return r.ResponseWriter.Write(b)
}

func sanitizeQueryParams(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ""
	}

	result := make(map[string]any, len(values))
	for key, val := range values {
		if isSensitiveField(key) {
			result[key] = "[REDACTED]"
			continue
		}
		result[key] = strings.Join(val, ", ")
	}

	return formatSanitizedPayload(result)
}

func sanitizeResponseBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Sprintf("[omitted response body, %d bytes]", len(body))
	}

	redactSensitiveValue(payload)
	return formatSanitizedPayload(payload)
}

func formatSanitizedPayload(payload any) string {
	switch v := payload.(type) {
	case string:
		return v
	default:
		buf, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", payload)
		}
		return string(buf)
	}
}

func shouldOmitOperationPayload(path string) bool {
	switch path {
	case "/go-api/internal/admin/auth/token",
		"/go-api/internal/admin/auth/password/reset",
		"/go-api/internal/admin/auth/password",
		"/go-api/internal/admin/auth/identifier",
		"/go-api/internal/admin/auth/reauth/password",
		"/go-api/internal/admin/auth/reauth/totp",
		"/go-api/internal/admin/auth/oauth/bind/confirm",
		"/go-api/internal/admin/auth/passkey/register/finish",
		"/go-api/internal/admin/auth/passkey/login/finish",
		"/go-api/internal/admin/auth/reauth/passkey/finish",
		"/go-api/internal/admin/auth/tfa/enable",
		"/go-api/internal/admin/auth/tfa/disable",
		"/go-api/internal/admin/system/user/password/reset",
		"/go-api/internal/admin/system/user/tfa/disable",
		"/go-api/internal/admin/system/user/passkey",
		"/go-api/internal/admin/system/user/passkeys":
		return true
	default:
		return false
	}
}
