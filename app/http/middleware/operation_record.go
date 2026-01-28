package middleware

import (
	"bytes"
	"context"
	"io"
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
		ctx := context.WithValue(context.Background(), logger.TraceIDKey, traceID)

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
	switch {
	case strings.Contains(c.GetHeader("Content-Type"), "boundary"):
		return `{"msg":"upload file"}`
	case c.Request.ContentLength > 65535:
		return `{"msg":"request body too large"}`
	case c.Request.Method == "GET":
		return c.Request.URL.RawQuery
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
	return string(body)
}

// createOperationRecord creates an operation record.
func (m middleware) createOperationRecord(c *gin.Context, params, traceID string, startTime time.Time, writer *responseBodyWriter) *system.OperationRecord {
	responseBody := ""
	if c.Request.Method != "GET" {
		responseBody = writer.body.String()
	}

	record := &system.OperationRecord{
		IP:           util.GetRealIP(c),
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		Agent:        c.Request.UserAgent(),
		Params:       params,
		CreateAt:     startTime,
		Status:       c.Writer.Status(),
		Latency:      time.Since(startTime).Seconds(),
		Resp:         responseBody,
		ErrorMessage: writer.errMessage,
		TraceID:      traceID,
	}

	if userID, exists := c.Get("user_id"); exists {
		record.UserID = userID.(uint)
	}
	if userName, exists := c.Get("user_name"); exists {
		record.UserName = userName.(string)
	}

	return record
}

// responseBodyWriter is a custom response body writer.
type responseBodyWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	logger     *logger.Manager
	errMessage string
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
	if code := jsoniter.Get(b, "code").ToInt(); code != 0 {
		r.errMessage = jsoniter.Get(b, "msg").ToString()
	}
	return r.ResponseWriter.Write(b)
}
