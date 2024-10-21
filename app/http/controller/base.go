package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
)

// BaseController Base Controller
type BaseController struct {
	AppCtx *http.Context
	Logger *logger.Manager
	Redis  *redis.Manager
	I18n   *i18n.Manager
}

// Context creates a new context with the trace ID from the gin.Context.
//
// Parameters:
//   - c: *gin.Context - The gin context containing the trace ID.
//
// Returns:
//   - context.Context: A new context with the trace ID added.
func (b *BaseController) Context(c *gin.Context) context.Context {
	return b.AppCtx.Context(c)
}
