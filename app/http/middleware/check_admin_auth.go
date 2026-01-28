package middleware

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/util"
	"strings"
)

// CheckAdminAuth returns a middleware function that validates admin authentication
// and checks user permissions before allowing access to protected routes.
func (m middleware) CheckAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := getContext(c)
		userID, errCode, err := m.checkAdminAuthToken(c)
		if errCode != e.SUCCESS {
			m.i18n.JSON(c, errCode, nil, err)
			c.Abort()
			return
		}

		ok, err := m.authService.HasRole(ctx, userID, "super_admin")
		if !ok {
			uri := c.Request.RequestURI
			permissionHash := util.MD5(c.Request.Method + strings.Split(uri, "?")[0])

			ok, err = m.authService.HasPermission(ctx, userID, permissionHash)
			if !ok {
				m.i18n.JSON(c, e.AccountInsufficientPermissions, nil, err)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func getContext(c *gin.Context) context.Context {
	traceID, ok := c.Get("trace_id")
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), logger.TraceIDKey, traceID.(string))
}

func (m middleware) checkAdminAuthToken(c *gin.Context) (userID uint, errCode int, err error) {
	errCode = e.ServerUnauthorized
	// First, try to get the token from the Authorization header
	token := c.Request.Header.Get("Authorization")

	// If not found, try to get it from the Cookie
	if token == "" {
		if cookie, cookieErr := c.Request.Cookie("admin-token"); cookieErr == nil {
			token = cookie.Value
		}
	}

	if token != "" {
		var userName string

		errCode = e.SUCCESS
		userName, userID, err = m.authService.VerifyToken(token)
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				errCode = e.AccountAuthorizationExpired
			case errors.Is(err, jwt.ErrTokenMalformed):
				errCode = e.InvalidAccount
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				errCode = e.TokenSignatureInvalid
			default:
				errCode = e.AccountUnauthorized
			}

			return
		}

		c.Set("user_id", userID)
		c.Set("user_name", userName)
	}

	return
}
