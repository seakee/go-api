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
