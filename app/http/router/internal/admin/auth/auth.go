package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/admin/auth"
)

func registerAuthRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := auth.NewHandler(ctx)
	rateLimitedAPI := api.Group("", ctx.Middleware.AdminAuthRateLimit())

	api.GET("oauth/url", handler.OAuthUrl())
	rateLimitedAPI.POST("token", handler.Token())
	rateLimitedAPI.POST("reauth", handler.Reauth())
	rateLimitedAPI.POST("oauth/bind/confirm", handler.ConfirmOAuthBind())
	api.GET("oauth/accounts", ctx.Middleware.CheckAdminAuth(), handler.OAuthAccounts())
	rateLimitedAPI.POST("oauth/unbind", ctx.Middleware.CheckAdminAuth(), handler.UnbindOAuth())
	api.GET("profile", ctx.Middleware.CheckAdminAuth(), handler.Profile())
	rateLimitedAPI.PUT("password/reset", handler.ResetPassword())
	rateLimitedAPI.PUT("password", ctx.Middleware.CheckAdminAuth(), handler.UpdatePassword())
	api.PUT("profile", ctx.Middleware.CheckAdminAuth(), handler.UpdateProfile())
	api.GET("menus", ctx.Middleware.CheckAdminAuth(), handler.UserMenuList())
	rateLimitedAPI.PUT("identifier", ctx.Middleware.CheckAdminAuth(), handler.UpdateIdentifier())
	rateLimitedAPI.PUT("tfa/enable", ctx.Middleware.CheckAdminAuth(), handler.EnableTfa())
	rateLimitedAPI.PUT("tfa/disable", ctx.Middleware.CheckAdminAuth(), handler.DisableTfa())
	api.GET("tfa/key", ctx.Middleware.CheckAdminAuth(), handler.TotpKey())
	api.GET("tfa/status", ctx.Middleware.CheckAdminAuth(), handler.TfaStatus())
}
