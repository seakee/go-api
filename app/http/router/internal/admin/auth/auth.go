package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/admin/auth"
)

func registerAuthRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := auth.NewHandler(ctx)

	api.GET("oauth/url", handler.OAuthUrl())
	api.POST("token", handler.Token())
	api.GET("profile", ctx.Middleware.CheckAdminAuth(), handler.Profile())
	api.PUT("password/reset", handler.ResetPassword())
	api.PUT("password", ctx.Middleware.CheckAdminAuth(), handler.UpdatePassword())
	api.PUT("profile", ctx.Middleware.CheckAdminAuth(), handler.UpdateProfile())
	api.GET("menus", ctx.Middleware.CheckAdminAuth(), handler.UserMenuList())
	api.PUT("account", ctx.Middleware.CheckAdminAuth(), handler.UpdateAccount())
	api.PUT("tfa/enable", ctx.Middleware.CheckAdminAuth(), handler.EnableTfa())
	api.PUT("tfa/disable", ctx.Middleware.CheckAdminAuth(), handler.DisableTfa())
	api.GET("tfa/key", ctx.Middleware.CheckAdminAuth(), handler.TotpKey())
	api.GET("tfa/status", ctx.Middleware.CheckAdminAuth(), handler.TfaStatus())
}
