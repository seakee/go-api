package system

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
)

func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
	api.GET("ping", func(c *gin.Context) {
		ctx.I18n.JSON(c, 0, nil, nil)
	})

	menuAPI := api.Group("menu")
	registerMenuRoutes(menuAPI, ctx)

	permissionAPI := api.Group("permission")
	registerPermissionRoutes(permissionAPI, ctx)

	roleAPI := api.Group("role")
	registerRoleRoutes(roleAPI, ctx)

	userAPI := api.Group("user")
	registerUserRoutes(userAPI, ctx)

	recordAPI := api.Group("record")
	registerRecordRoutes(recordAPI, ctx)
}
