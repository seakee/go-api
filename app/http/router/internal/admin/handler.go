package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/router/internal/admin/auth"
	"github.com/seakee/go-api/app/http/router/internal/admin/system"
)

func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
	api.GET("ping", func(c *gin.Context) {
		ctx.I18n.JSON(c, 0, nil, nil)
	})

	systemAPI := api.Group("system", ctx.Middleware.CheckAdminAuth())
	system.RegisterRoutes(systemAPI, ctx)

	authAPI := api.Group("auth")
	auth.RegisterRoutes(authAPI, ctx)
}
