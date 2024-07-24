// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package router handles the routing for the application.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http/controller/auth"
)

// authGroup sets up the routes for authentication-related endpoints.
//
// Parameters:
//   - api: *gin.RouterGroup - The router group to add the auth routes to.
//   - core: *Core - The core application structure containing necessary dependencies.
func authGroup(api *gin.RouterGroup, core *Core) {
	// Create a new auth handler
	authHandler := auth.New(core.Logger, core.Redis["go-api"], core.I18n, core.MysqlDB["go-api"])
	{
		// POST /app - Create a new app (requires app authentication)
		api.POST("app", core.Middleware.CheckAppAuth(), authHandler.Create())
		// POST /token - Get a new token
		api.POST("token", authHandler.GetToken())
	}
}
