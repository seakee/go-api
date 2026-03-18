package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	authController "github.com/seakee/go-api/app/http/controller/admin/auth"
)

// registerAuthRoutes organizes admin auth endpoints by shared middleware and business domain.
func registerAuthRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := authController.NewHandler(ctx)
	// Public auth endpoints that should be throttled before reaching the handler.
	rateLimitedAPI := api.Group("", ctx.Middleware.AdminAuthRateLimit())
	// Authenticated endpoints that do not need additional brute-force protection.
	protectedAPI := api.Group("", ctx.Middleware.CheckAdminAuth())
	// Sensitive authenticated endpoints keep the original order: rate limit first, auth check second.
	protectedRateLimitedAPI := rateLimitedAPI.Group("", ctx.Middleware.CheckAdminAuth())

	registerSessionRoutes(api, rateLimitedAPI, handler)
	registerReauthRoutes(rateLimitedAPI, protectedAPI, protectedRateLimitedAPI, handler)
	registerOAuthRoutes(rateLimitedAPI, protectedAPI, protectedRateLimitedAPI, handler)
	registerPasskeyRoutes(rateLimitedAPI, protectedAPI, handler)
	registerAccountRoutes(rateLimitedAPI, protectedAPI, protectedRateLimitedAPI, handler)
}

// Session and primary login flows.
func registerSessionRoutes(
	api *gin.RouterGroup,
	rateLimitedAPI *gin.RouterGroup,
	handler authController.Handler,
) {
	api.GET("oauth/url", handler.OAuthUrl())
	rateLimitedAPI.POST("token", handler.Token())
}

// Reauthentication flows used before sensitive follow-up operations.
func registerReauthRoutes(
	rateLimitedAPI *gin.RouterGroup,
	protectedAPI *gin.RouterGroup,
	protectedRateLimitedAPI *gin.RouterGroup,
	handler authController.Handler,
) {
	rateLimitedAPI.POST("reauth", handler.Reauth())
	protectedAPI.GET("reauth/methods", handler.ReauthMethods())
	protectedRateLimitedAPI.POST("reauth/password", handler.ReauthByPassword())
	protectedRateLimitedAPI.POST("reauth/totp", handler.ReauthByTotp())
	protectedRateLimitedAPI.POST("reauth/passkey/options", handler.BeginPasskeyReauth())
	protectedRateLimitedAPI.POST("reauth/passkey/finish", handler.FinishPasskeyReauth())
}

// OAuth login binding and account management endpoints.
func registerOAuthRoutes(
	rateLimitedAPI *gin.RouterGroup,
	protectedAPI *gin.RouterGroup,
	protectedRateLimitedAPI *gin.RouterGroup,
	handler authController.Handler,
) {
	rateLimitedAPI.POST("oauth/bind/confirm", handler.ConfirmOAuthBind())
	protectedAPI.GET("oauth/accounts", handler.OAuthAccounts())
	protectedRateLimitedAPI.POST("oauth/unbind", handler.UnbindOAuth())
}

// Passkey login, registration, and credential management endpoints.
func registerPasskeyRoutes(
	rateLimitedAPI *gin.RouterGroup,
	protectedAPI *gin.RouterGroup,
	handler authController.Handler,
) {
	rateLimitedAPI.POST("passkey/login/options", handler.BeginPasskeyLogin())
	rateLimitedAPI.POST("passkey/login/finish", handler.FinishPasskeyLogin())
	protectedAPI.POST("passkey/register/options", handler.BeginPasskeyRegistration())
	protectedAPI.POST("passkey/register/finish", handler.FinishPasskeyRegistration())
	protectedAPI.GET("passkeys", handler.Passkeys())
	protectedAPI.DELETE("passkey", handler.DeletePasskey())
}

// Profile, password, identifier, menu, and TFA endpoints.
func registerAccountRoutes(
	rateLimitedAPI *gin.RouterGroup,
	protectedAPI *gin.RouterGroup,
	protectedRateLimitedAPI *gin.RouterGroup,
	handler authController.Handler,
) {
	protectedAPI.GET("profile", handler.Profile())
	rateLimitedAPI.PUT("password/reset", handler.ResetPassword())
	protectedRateLimitedAPI.PUT("password", handler.UpdatePassword())
	protectedAPI.PUT("profile", handler.UpdateProfile())
	protectedAPI.GET("menus", handler.UserMenuList())
	protectedRateLimitedAPI.PUT("identifier", handler.UpdateIdentifier())
	protectedRateLimitedAPI.PUT("tfa/enable", handler.EnableTfa())
	protectedRateLimitedAPI.PUT("tfa/disable", handler.DisableTfa())
	protectedAPI.GET("tfa/key", handler.TotpKey())
	protectedAPI.GET("tfa/status", handler.TfaStatus())
}
