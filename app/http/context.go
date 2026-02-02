package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/seakee/go-api/app/config"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/notify"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Context http context
type Context struct {
	Logger        *logger.Manager
	Redis         map[string]*redis.Manager
	I18n          *i18n.Manager
	SqlDB         map[string]*gorm.DB
	MongoDB       map[string]*qmgo.Database
	Middleware    middleware.Middleware
	KafkaProducer *kafka.Manager
	Notify        *notify.Manager
	Config        *config.Config
	Engine        *gin.Engine
}

// Context creates a new context with the trace ID from the gin.Context.
//
// Parameters:
//   - c: *gin.Context - The gin context containing the trace ID.
//
// Returns:
//   - context.Context: A new context with the trace ID added.
func (ctx *Context) Context(c *gin.Context) context.Context {
	traceID, ok := c.Get("trace_id")
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), logger.TraceIDKey, traceID.(string))
}
