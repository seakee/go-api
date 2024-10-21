// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/auth"
)

// RegisterRoutes sets up the routes for authentication-related endpoints.
//
// Parameters:
//   - api: *gin.RouterGroup - The router group to add the auth routes to.
//   - core: *Core - The core application structure containing necessary dependencies.
func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
	api.GET("ping", func(c *gin.Context) {
		ctx.I18n.JSON(c, 0, nil, nil)
	})

	// Create a new auth handler
	authHandler := auth.NewHandler(ctx)
	{
		// POST /app - Create a new app (requires app authentication)
		api.POST("app", ctx.Middleware.CheckAppAuth(), authHandler.Create())
		// POST /token - Get a new token
		api.POST("token", authHandler.GetToken())
	}
}
