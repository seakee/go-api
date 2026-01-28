package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
)

func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
	api.GET("ping", func(c *gin.Context) {
		ctx.I18n.JSON(c, 0, nil, nil)
	})

	registerAuthRoutes(api, ctx)
}
