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
// 2. Validates the app credentials against the database.
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

		ctx := h.ctx(c)

		errCode = e.InvalidParams
		if appID != "" && appSecret != "" {
			// Attempt to retrieve the app from the database
			app, err = h.repo.GetApp(ctx, &auth.App{AppID: appID, AppSecret: appSecret, Status: 1})
			errCode = e.ServerAppNotFound
			if err == nil {
				// Generate a JWT token for the app
				token, err = jwt.GenerateAppToken(app, appTokenExpireTime)
				errCode = e.ServerAuthorizationFail
				if err == nil {
					errCode = e.SUCCESS
					data["token"] = token
					data["expires_in"] = appTokenExpireTime
				}
			}
		}

		// Respond with the result
		h.i18n.JSON(c, errCode, data, err)
	}
}
