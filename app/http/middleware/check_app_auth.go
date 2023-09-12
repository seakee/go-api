package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app/pkg/e"
	apiJWT "github.com/seakee/go-api/app/pkg/jwt"
)

func (m middleware) CheckAppAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		errCode, err := checkByToken(c)
		if errCode != e.SUCCESS {
			m.i18n.JSON(c, errCode, nil, err)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkByToken 返回通过token验证的结果
func checkByToken(c *gin.Context) (errCode int, err error) {
	errCode = e.InvalidParams

	token := c.Request.Header.Get("Authorization")
	if token != "" {
		var serverClaims *apiJWT.ServerClaims

		errCode = e.SUCCESS

		serverClaims, err = apiJWT.ParseAppAuth(token)
		if err != nil {
			switch err {
			case jwt.ErrTokenExpired:
				errCode = e.ServerAuthorizationExpired
			default:
				errCode = e.ServerUnauthorized
			}
		} else {
			c.Set("app_id", serverClaims.AppID)
			c.Set("app_name", serverClaims.AppName)
		}
	}

	return
}
