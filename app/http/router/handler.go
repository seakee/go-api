// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package router handles the routing for the application.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/router/external"
	"github.com/seakee/go-api/app/http/router/internal"
)

func Register(engine *gin.Engine, ctx *http.Context) {
	ctx.Engine = engine

	routePrefix := "go-api"
	if ctx != nil && ctx.Config != nil && ctx.Config.System.RoutePrefix != "" {
		routePrefix = ctx.Config.System.RoutePrefix
	}

	api := engine.Group(routePrefix)

	// Set up internal API routes
	internalAPI := api.Group("internal")
	internal.RegisterRoutes(internalAPI, ctx)

	// Set up external API routes
	externalAPI := api.Group("external")
	external.RegisterRoutes(externalAPI, ctx)
}
