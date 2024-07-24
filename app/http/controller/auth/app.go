// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package auth provides authentication-related functionality for the application.
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/model/auth"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/sk-pkg/util"
)

// StoreAppReqParams defines the structure for storing app request parameters.
type StoreAppReqParams struct {
	AppName     string `json:"app_name" form:"app_name" binding:"required"`
	Description string `json:"description" form:"description"`
	RedirectUri string `json:"redirect_uri" form:"redirect_uri"`
}

// StoreAppRepData defines the structure for the response data when storing an app.
type StoreAppRepData struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// Create returns a gin.HandlerFunc that handles the creation of a new app.
//
// This function performs the following steps:
// 1. Binds the JSON request to StoreAppReqParams.
// 2. Checks if an app with the given name already exists.
// 3. If the app doesn't exist, creates a new app with generated AppID and AppSecret.
// 4. Returns the newly created AppID and AppSecret.
//
// Returns:
//   - gin.HandlerFunc: A function that handles the HTTP request for creating an app.
func (h handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params *StoreAppReqParams
		var err error
		var exists bool
		var data *StoreAppRepData

		errCode := e.InvalidParams

		ctx := h.ctx(c)

		// Bind JSON request to StoreAppReqParams
		if err = c.ShouldBindJSON(&params); err == nil {
			// Check if app already exists
			exists, err = h.repo.ExistAppByName(ctx, params.AppName)
			errCode = e.ServerAppAlreadyExists
			if !exists {
				// Create new app
				app := &auth.App{
					AppName:     params.AppName,
					AppID:       "go-api-" + util.RandLowStr(8),
					AppSecret:   util.RandUpStr(32),
					RedirectUri: params.RedirectUri,
					Description: params.Description,
					Status:      1,
				}

				// Save app to repository
				_, err = h.repo.Create(ctx, app)
				errCode = e.BUSY
				if err == nil {
					errCode = e.SUCCESS

					// Prepare response data
					data = &StoreAppRepData{
						AppID:     app.AppID,
						AppSecret: app.AppSecret,
					}
				}
			}
		}

		// Send JSON response
		h.i18n.JSON(c, errCode, data, err)
	}
}
