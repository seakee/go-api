package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/repository/auth"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	Handler interface {
		i()
		Create() gin.HandlerFunc
		GetToken() gin.HandlerFunc
	}

	handler struct {
		logger *zap.Logger
		redis  *redis.Manager
		i18n   *i18n.Manager
		repo   auth.Repo
	}
)

func (h handler) i() {}

func New(logger *zap.Logger, redis *redis.Manager, i18n *i18n.Manager, db *gorm.DB) Handler {
	return &handler{
		logger: logger,
		redis:  redis,
		i18n:   i18n,
		repo:   auth.NewAppRepo(db, redis),
	}
}
