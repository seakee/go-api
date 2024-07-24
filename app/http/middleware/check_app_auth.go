// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package middleware provides HTTP middleware functions for the application.
package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app/pkg/e"
	apiJWT "github.com/seakee/go-api/app/pkg/jwt"
)

// CheckAppAuth returns a Gin middleware function that checks the application's authentication.
//
// This middleware validates the JWT token in the "Authorization" header.
// If the token is valid, it sets the app_id and app_name in the Gin context.
// If the token is invalid or expired, it aborts the request with an appropriate error response.
//
// Returns:
//   - gin.HandlerFunc: A middleware function for Gin framework.
func (m middleware) CheckAppAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		errCode, err := checkByToken(c)
		if errCode != e.SUCCESS {
			// If authentication fails, respond with an error and abort the request
			m.i18n.JSON(c, errCode, nil, err)
			c.Abort()
			return
		}

		// If authentication succeeds, continue to the next middleware or handler
		c.Next()
	}
}

// checkByToken validates the JWT token from the request header and returns the result.
//
// It extracts the token from the "Authorization" header, parses it, and sets
// app_id and app_name in the Gin context if the token is valid.
//
// Parameters:
//   - c: *gin.Context - The Gin context containing the HTTP request information.
//
// Returns:
//   - errCode: int - An error code indicating the result of the token validation.
//   - err: error - An error object if the token validation fails, or nil if successful.
//
// Error codes:
//   - e.SUCCESS: Token is valid
//   - e.InvalidParams: No token provided
//   - e.ServerAuthorizationExpired: Token has expired
//   - e.ServerUnauthorized: Token is invalid
func checkByToken(c *gin.Context) (errCode int, err error) {
	errCode = e.InvalidParams

	// Extract token from the Authorization header
	token := c.Request.Header.Get("Authorization")
	if token != "" {
		var serverClaims *apiJWT.ServerClaims

		errCode = e.SUCCESS

		// Parse and validate the token
		serverClaims, err = apiJWT.ParseAppAuth(token)
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				errCode = e.ServerAuthorizationExpired
			default:
				errCode = e.ServerUnauthorized
			}
		} else {
			// If token is valid, set app_id and app_name in the context
			c.Set("app_id", serverClaims.AppID)
			c.Set("app_name", serverClaims.AppName)
		}
	}

	return
}
