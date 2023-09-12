package router

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http/controller/auth"
)

func authGroup(api *gin.RouterGroup, core *Core) {
	authHandler := auth.New(core.Logger, core.Redis["go-api"], core.I18n, core.MysqlDB["go_api"])
	{
		api.POST("app", core.Middleware.CheckAppAuth(), authHandler.Create())
		api.POST("token", authHandler.GetToken())
	}
}
