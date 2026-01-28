package system

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/admin/system/role"
)

func registerRoleRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := role.NewHandler(ctx)

	api.POST("", handler.Create())
	api.DELETE("", handler.Delete())
	api.PUT("", handler.Update())
	api.GET("", handler.Detail())
	api.GET("list", handler.List())
	api.GET("paginate", handler.Paginate())

	api.GET("permission", handler.Permissions())
	api.PUT("permission", handler.UpdatePermission())
}
