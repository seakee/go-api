// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/model/auth"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/pkg/jwt"
)

// appTokenExpireTime defines the expiration time for app tokens in seconds (7 days).
const appTokenExpireTime = 168 * 3600

// GetToken is a gin.HandlerFunc that generates and returns an authentication token for an app.
//
// This function handles the following steps:
// 1. Extracts app_id and app_secret from the POST form data.
// 2. Validates the app credentials using the service layer.
// 3. Generates a JWT token if the credentials are valid.
// 4. Returns the token and its expiration time, or an error if any step fails.
//
// Returns:
//   - gin.HandlerFunc: A function that can be used as a Gin route handler.
func (h handler) GetToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			err     error
			app     *auth.App
			errCode int
			data    gin.H
			token   string
		)

		// Extract app credentials from the POST form
		appID := c.PostForm("app_id")
		appSecret := c.PostForm("app_secret")
		data = make(gin.H)

		errCode = e.InvalidParams
		if appID != "" && appSecret != "" {
			// Use service layer to validate credentials
			app, err = h.service.GetTokenByCredentials(h.Context(c), appID, appSecret)
			if err != nil {
				// Handle specific service errors
				if err.Error() == "invalid credentials" {
					errCode = e.ServerAppNotFound
				} else if err.Error() == "app is not active" {
					errCode = e.ServerAppNotFound
				} else {
					errCode = e.ServerAuthorizationFail
				}
			} else {
				// Generate a JWT token for the app
				token, err = jwt.GenerateAppToken(app, appTokenExpireTime)
				if err != nil {
					errCode = e.ServerAuthorizationFail
				} else {
					errCode = e.SUCCESS
					data["token"] = token
					data["expires_in"] = appTokenExpireTime
				}
			}
		}

		// Respond with the result
		h.I18n.JSON(c, errCode, data, err)
	}
}
