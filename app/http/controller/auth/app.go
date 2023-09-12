package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/model/auth"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/sk-pkg/util"
)

type (
	StoreAppReqParams struct {
		AppName     string `json:"app_name" form:"app_name" binding:"required"`
		Description string `json:"description" form:"description"`
		RedirectUri string `json:"redirect_uri" form:"redirect_uri"`
	}

	StoreAppRepData struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}
)

func (h handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params *StoreAppReqParams
		var err error
		var exists bool
		var data *StoreAppRepData

		errCode := e.InvalidParams

		if err = c.ShouldBindJSON(&params); err == nil {
			exists, err = h.repo.ExistAppByName(params.AppName)
			errCode = e.ServerAppAlreadyExists
			if !exists {
				app := &auth.App{
					AppName:     params.AppName,
					AppID:       "go-api-" + util.RandLowStr(8),
					AppSecret:   util.RandUpStr(32),
					RedirectUri: params.RedirectUri,
					Description: params.Description,
					Status:      1,
				}

				_, err = h.repo.Create(app)
				errCode = e.BUSY
				if err == nil {
					errCode = e.SUCCESS

					data = &StoreAppRepData{
						AppID:     app.AppID,
						AppSecret: app.AppSecret,
					}
				}
			}
		}

		h.i18n.JSON(c, errCode, data, err)
	}
}
