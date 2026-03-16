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
	rateLimitedAPI.POST("passkey/login/options", handler.BeginPasskeyLogin())
	rateLimitedAPI.POST("passkey/login/finish", handler.FinishPasskeyLogin())
	rateLimitedAPI.POST("reauth", handler.Reauth())
	api.GET("reauth/methods", ctx.Middleware.CheckAdminAuth(), handler.ReauthMethods())
	rateLimitedAPI.POST("reauth/password", ctx.Middleware.CheckAdminAuth(), handler.ReauthByPassword())
	rateLimitedAPI.POST("reauth/totp", ctx.Middleware.CheckAdminAuth(), handler.ReauthByTotp())
	rateLimitedAPI.POST("reauth/passkey/options", ctx.Middleware.CheckAdminAuth(), handler.BeginPasskeyReauth())
	rateLimitedAPI.POST("reauth/passkey/finish", ctx.Middleware.CheckAdminAuth(), handler.FinishPasskeyReauth())
	rateLimitedAPI.POST("oauth/bind/confirm", handler.ConfirmOAuthBind())
	api.GET("oauth/accounts", ctx.Middleware.CheckAdminAuth(), handler.OAuthAccounts())
	rateLimitedAPI.POST("oauth/unbind", ctx.Middleware.CheckAdminAuth(), handler.UnbindOAuth())
	api.POST("passkey/register/options", ctx.Middleware.CheckAdminAuth(), handler.BeginPasskeyRegistration())
	api.POST("passkey/register/finish", ctx.Middleware.CheckAdminAuth(), handler.FinishPasskeyRegistration())
	api.GET("passkeys", ctx.Middleware.CheckAdminAuth(), handler.Passkeys())
	api.DELETE("passkey", ctx.Middleware.CheckAdminAuth(), handler.DeletePasskey())
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
