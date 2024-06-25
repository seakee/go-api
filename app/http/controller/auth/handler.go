package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/repository/auth"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

type (
	Handler interface {
		i()
		ctx(c *gin.Context) context.Context
		Create() gin.HandlerFunc
		GetToken() gin.HandlerFunc
	}

	handler struct {
		logger *logger.Manager
		redis  *redis.Manager
		i18n   *i18n.Manager
		repo   auth.Repo
	}
)

func (h handler) ctx(c *gin.Context) context.Context {
	traceID, _ := c.Get("trace_id")

	return context.WithValue(context.Background(), logger.TraceIDKey, traceID.(string))
}

func (h handler) i() {}

func New(logger *logger.Manager, redis *redis.Manager, i18n *i18n.Manager, db *gorm.DB) Handler {
	return &handler{
		logger: logger,
		redis:  redis,
		i18n:   i18n,
		repo:   auth.NewAppRepo(db, redis),
	}
}
