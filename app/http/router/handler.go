// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

type Core struct {
	Logger        *logger.Manager
	Redis         map[string]*redis.Manager
	I18n          *i18n.Manager
	MysqlDB       map[string]*gorm.DB
	MongoDB       map[string]*qmgo.Database
	Middleware    middleware.Middleware
	KafkaProducer *kafka.Manager
}

func New(mux *gin.Engine, core *Core) *gin.Engine {
	api := mux.Group("go-api")
	// 内部调用接口
	internal(api.Group("internal"), core)
	// 外部调用接口
	external(api.Group("external"), core)

	return mux
}

// external 外部调用接口
func external(api *gin.RouterGroup, core *Core) {
	api.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// app接口群组
	appGroup := api.Group("app")
	appGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// 服务接口群组
	serviceGroup := api.Group("service")
	serviceGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})
}

// internal 内部调用接口
func internal(api *gin.RouterGroup, core *Core) {
	api.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// 管理平台接口群组
	adminGroup := api.Group("admin")
	adminGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// 服务接口群组
	serviceGroup := api.Group("service")
	serviceGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	authGroup(serviceGroup.Group("server/auth"), core)
}
