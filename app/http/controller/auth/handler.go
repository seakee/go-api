// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package auth provides authentication-related functionality for the application.
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	authService "github.com/seakee/go-api/app/service/auth"
)

// Handler interface defines the methods that should be implemented by the auth handler.
type Handler interface {
	i()
	Create() gin.HandlerFunc
	GetToken() gin.HandlerFunc
}

// handler struct implements the Handler interface.
type handler struct {
	controller.BaseController
	service authService.AppService
}

// i is a dummy method to satisfy the Handler interface.
func (h handler) i() {}

// NewHandler creates and returns a new Handler instance.
//
// Parameters:
//   - appCtx: *http.Context - The application context.
//
// Returns:
//   - Handler: A new Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: authService.NewAppService(appCtx.MysqlDB["go-api"], appCtx.Redis["go-api"]),
	}
}
