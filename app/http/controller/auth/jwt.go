package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/model/auth"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/pkg/jwt"
)

const appTokenExpireTime = 168 * 3600

func (h handler) GetToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			err     error
			app     *auth.App
			errCode int
			data    gin.H
			token   string
		)

		appID := c.PostForm("app_id")
		appSecret := c.PostForm("app_secret")
		data = make(gin.H)

		errCode = e.InvalidParams
		if appID != "" && appSecret != "" {
			app, err = h.repo.GetApp(&auth.App{AppID: appID, AppSecret: appSecret, Status: 1})
			errCode = e.ServerAppNotFound
			if err == nil {
				token, err = jwt.GenerateAppToken(app, appTokenExpireTime)
				errCode = e.ServerAuthorizationFail
				if err == nil {
					errCode = e.SUCCESS
					data["token"] = token
					data["expires_in"] = appTokenExpireTime
				}
			}
		}

		h.i18n.JSON(c, errCode, data, err)
	}
}
