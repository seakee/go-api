// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package middleware provides HTTP middleware functions for the application.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors returns a Gin middleware function that handles Cross-Origin Resource Sharing (CORS).
//
// This middleware sets the necessary CORS headers to allow cross-origin requests.
// It handles preflight requests (OPTIONS) by responding with a 204 No Content status.
//
// Returns:
//   - gin.HandlerFunc: A middleware function for Gin framework.
func (m middleware) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			// Set CORS headers
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "false")
			c.Set("content-type", "application/json")
		}

		// Handle preflight requests
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		// Continue to the next middleware or handler
		c.Next()
	}
}
