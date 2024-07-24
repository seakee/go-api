// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package e defines error codes and messages used throughout the go-api project.
// These codes help standardize error handling and client-side error interpretation.
package e

// Error codes
const (
	BUSY    = -1  // System is busy
	SUCCESS = 0   // Operation successful
	ERROR   = 500 // General server error

	InvalidParams = 400 // Invalid parameters

	ServerUnauthorized         = 10001 // Server is not authorized
	ServerAuthorizationExpired = 10002 // Server authorization has expired
	ServerAuthorizationFail    = 10003 // Server authorization failed
	ServerAppNotFound          = 10004 // Server application not found
	ServerAppAlreadyExists     = 10005 // Server application already exists
	ServerAPIUserNotFound      = 10006 // Server API user not found
	InvalidServerAppID         = 10007 // Invalid server application ID
)
