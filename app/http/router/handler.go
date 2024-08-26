// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package router handles the routing for the application.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"github.com/seakee/go-api/app/http/middleware"
	"github.com/sk-pkg/i18n"
	"github.com/sk-pkg/kafka"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/notify"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// Core struct holds all the core dependencies for the application.
type Core struct {
	Logger        *logger.Manager
	Redis         map[string]*redis.Manager
	I18n          *i18n.Manager
	MysqlDB       map[string]*gorm.DB
	MongoDB       map[string]*qmgo.Database
	Middleware    middleware.Middleware
	KafkaProducer *kafka.Manager
	Notify        *notify.Manager
}

// New sets up the main router for the application.
//
// Parameters:
//   - mux: *gin.Engine - The main Gin engine.
//   - core: *Core - The core application structure containing necessary dependencies.
//
// Returns:
//   - *gin.Engine: The configured Gin engine with all routes set up.
func New(mux *gin.Engine, core *Core) *gin.Engine {
	api := mux.Group("go-api")
	// Set up internal API routes
	internal(api.Group("internal"), core)
	// Set up external API routes
	external(api.Group("external"), core)

	return mux
}

// external sets up the routes for external API calls.
//
// Parameters:
//   - api: *gin.RouterGroup - The router group for external API routes.
//   - core: *Core - The core application structure containing necessary dependencies.
func external(api *gin.RouterGroup, core *Core) {
	// GET /ping - Health check endpoint
	api.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// Set up app-related routes
	appGroup := api.Group("app")
	appGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// Set up service-related routes
	serviceGroup := api.Group("service")
	serviceGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})
}

// internal sets up the routes for internal API calls.
//
// Parameters:
//   - api: *gin.RouterGroup - The router group for internal API routes.
//   - core: *Core - The core application structure containing necessary dependencies.
func internal(api *gin.RouterGroup, core *Core) {
	// GET /ping - Health check endpoint
	api.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// Set up admin-related routes
	adminGroup := api.Group("admin")
	adminGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// Set up service-related routes
	serviceGroup := api.Group("service")
	serviceGroup.GET("ping", func(c *gin.Context) {
		core.I18n.JSON(c, 0, nil, nil)
	})

	// Set up authentication-related routes
	authGroup(serviceGroup.Group("server/auth"), core)
}
