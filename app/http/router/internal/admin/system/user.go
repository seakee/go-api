package system

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/admin/system/user"
)

func registerUserRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := user.NewHandler(ctx)

	api.POST("", handler.Create())
	api.DELETE("", handler.Delete())
	api.PUT("", handler.Update())
	api.GET("", handler.Detail())
	api.GET("paginate", handler.Paginate())

	api.GET("role", handler.Roles())
	api.PUT("role", handler.UpdateRole())
}
