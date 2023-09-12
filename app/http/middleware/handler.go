package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	Middleware interface {
		CheckAppAuth() gin.HandlerFunc

		Cors() gin.HandlerFunc

		RequestLogger() gin.HandlerFunc
	}

	middleware struct {
		logger *zap.Logger
		i18n   *i18n.Manager
		db     map[string]*gorm.DB
		redis  map[string]*redis.Manager
	}
)

func New(logger *zap.Logger, i18n *i18n.Manager, db map[string]*gorm.DB, redis map[string]*redis.Manager) Middleware {
	return &middleware{logger: logger, i18n: i18n, db: db, redis: redis}
}
