// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors returns a Gin middleware that enforces a strict CORS allowlist.
//
// Security behavior:
//  1. Requests are allowed only when the Origin header exists and matches
//     CORS_ALLOW_ORIGINS (comma-separated list, case-insensitive).
//  2. When an origin is allowed, the middleware reflects that exact origin
//     and sets standard CORS response headers.
//  3. Preflight (OPTIONS) requests from disallowed origins are rejected with
//     HTTP 403 to avoid permissive cross-origin probing.
//
// Operational notes:
//  - If CORS_ALLOW_ORIGINS is empty, cross-origin access is denied by default.
//  - Use CORS_ALLOW_ORIGINS="*" only when the deployment explicitly accepts
//    all origins.
//
// Returns:
//   - gin.HandlerFunc: middleware used by the Gin engine.
func (m middleware) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		// Allow only explicit origins configured by runtime environment.
		allowOrigin := origin != "" && isAllowedOrigin(origin)
		if allowOrigin {
			// Expose CORS headers only for trusted origins.
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "false")
			// Ensure caches separate responses by Origin when CORS is active.
			c.Header("Vary", "Origin")
			c.Set("content-type", "application/json")
		}

		// Handle CORS preflight requests before routing handlers run.
		if method == "OPTIONS" {
			if origin != "" && !allowOrigin {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Continue to the next middleware or handler
		c.Next()
	}
}

// isAllowedOrigin checks whether origin is present in CORS_ALLOW_ORIGINS.
//
// CORS_ALLOW_ORIGINS format:
//   - Comma-separated origins, e.g.:
//     "https://admin.example.com,https://portal.example.com"
//   - Optional wildcard "*" for explicitly open configurations.
//
// The comparison is case-insensitive and ignores surrounding spaces.
func isAllowedOrigin(origin string) bool {
	allowOrigin := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if allowOrigin == "" {
		return false
	}

	for _, o := range strings.Split(allowOrigin, ",") {
		candidate := strings.TrimSpace(o)
		if candidate == "" {
			continue
		}
		if candidate == "*" || strings.EqualFold(candidate, origin) {
			return true
		}
	}

	return false
}
