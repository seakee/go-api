package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/pkg/trace"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

type (
	Middleware interface {
		CheckAppAuth() gin.HandlerFunc

		Cors() gin.HandlerFunc

		RequestLogger() gin.HandlerFunc

		SetTraceID() gin.HandlerFunc
	}

	middleware struct {
		logger  *logger.Manager
		i18n    *i18n.Manager
		db      map[string]*gorm.DB
		redis   map[string]*redis.Manager
		traceID *trace.ID
	}
)

func New(logger *logger.Manager, i18n *i18n.Manager, db map[string]*gorm.DB, redis map[string]*redis.Manager, traceID *trace.ID) Middleware {
	return &middleware{logger: logger, i18n: i18n, db: db, redis: redis, traceID: traceID}
}
