// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package e

const (
	BUSY          = -1
	SUCCESS       = 0
	ERROR         = 500
	InvalidParams = 400

	ServerUnauthorized         = 10001
	ServerAuthorizationExpired = 10002
	ServerAuthorizationFail    = 10003
	ServerAppNotFound          = 10004
	ServerAppAlreadyExists     = 10005
	ServerAPIUserNotFound      = 10006
	InvalidServerAppID         = 10007
)
