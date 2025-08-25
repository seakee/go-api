// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/pkg/e"
	authService "github.com/seakee/go-api/app/service/auth"
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
// 2. Calls the service layer to handle the business logic.
// 3. Returns the result or error response.
//
// Returns:
//   - gin.HandlerFunc: A function that handles the HTTP request for creating an app.
func (h handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params StoreAppReqParams
		var err error
		var data *StoreAppRepData

		errCode := e.InvalidParams
		ctx := h.Context(c)

		// Bind JSON request to StoreAppReqParams
		if err = c.ShouldBindJSON(&params); err == nil {
			// Convert to service params
			serviceParams := &authService.CreateAppParams{
				AppName:     params.AppName,
				Description: params.Description,
				RedirectUri: params.RedirectUri,
			}

			// Call service layer
			result, err := h.service.CreateApp(ctx, serviceParams)
			if err != nil {
				// Handle specific business errors
				if err.Error() == "app with this name already exists" {
					errCode = e.ServerAppAlreadyExists
				} else {
					errCode = e.BUSY
				}
			} else {
				errCode = e.SUCCESS
				// Prepare response data
				data = &StoreAppRepData{
					AppID:     result.AppID,
					AppSecret: result.AppSecret,
				}
			}
		}

		// Send JSON response
		h.I18n.JSON(c, errCode, data, err)
	}
}
