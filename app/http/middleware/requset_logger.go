package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
	"io"
	"time"
)

// RequestLogger 记录请求日志
func (m middleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		buf, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := util.GetRealIP(c)

		m.logger.Info(
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
