package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
)

// Handler defines the interface for authentication-related HTTP handlers.
type Handler interface {
	Token() gin.HandlerFunc

	Profile() gin.HandlerFunc
	UpdateProfile() gin.HandlerFunc

	ResetPassword() gin.HandlerFunc
	UpdatePassword() gin.HandlerFunc

	OAuthUrl() gin.HandlerFunc

	UserMenuList() gin.HandlerFunc

	UpdateAccount() gin.HandlerFunc

	DisableTfa() gin.HandlerFunc
	EnableTfa() gin.HandlerFunc
	TotpKey() gin.HandlerFunc
	TfaStatus() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.AuthService
}

func (h handler) TfaStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		enable, errCode, err := h.service.TfaStatus(h.Context(c), userID.(uint))

		data := gin.H{"enable": enable}

		h.I18n.JSON(c, errCode, data, err)
	}
}

func (h handler) DisableTfa() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TotpCode string `json:"totp_code" binding:"required"`
		}

		var err error

		errCode := e.InvalidParams

		userID, _ := c.Get("user_id")

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.DisableTfa(h.Context(c), userID.(uint), req.TotpCode)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) EnableTfa() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TotpCode string `json:"totp_code" binding:"required"`
			TotpKey  string `json:"totp_key" binding:"required"`
		}

		var err error

		errCode := e.InvalidParams

		userID, _ := c.Get("user_id")

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.EnableTfa(h.Context(c), userID.(uint), req.TotpCode, req.TotpKey)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) TotpKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		key, qrCode, errCode, err := h.service.TotpKey(h.Context(c), userID.(uint))
		data := gin.H{"totp_key": key, "qr_code": qrCode}

		h.I18n.JSON(c, errCode, data, err)
	}
}

func (h handler) UpdateAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TotpCode string `json:"totp_code" binding:"required"`
			Account  string `json:"account" binding:"required"`
		}

		var err error

		errCode := e.InvalidParams

		userID, _ := c.Get("user_id")

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.UpdateAccount(h.Context(c), userID.(uint), req.TotpCode, req.Account)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) UserMenuList() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.Context(c)

		errCode := e.InvalidParams

		userid, _ := c.Get("user_id")

		list, err := h.service.UserMenuList(ctx, userid.(uint))
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, gin.H{"items": list}, err)
	}
}

func (h handler) Profile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.Context(c)

		errCode := e.InvalidParams

		userid, _ := c.Get("user_id")

		list, errCode, err := h.service.Profile(ctx, userid.(uint))
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, list, err)
	}
}

func (h handler) UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserName string `json:"user_name"`
			Avatar   string `json:"avatar"`
		}

		var err error

		userid, _ := c.Get("user_id")

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.UpdateProfile(h.Context(c), userid.(uint), req.UserName, req.Avatar)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SafeCode string `json:"safe_code" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var err error

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.ResetPassword(h.Context(c), req.SafeCode, req.Password)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TotpCode string `json:"totp_code" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var err error

		errCode := e.InvalidParams

		userID, _ := c.Get("user_id")

		if err = c.ShouldBind(&req); err == nil {
			errCode, err = h.service.UpdatePassword(h.Context(c), userID.(uint), req.TotpCode, req.Password)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) OAuthUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OauthType string `form:"type" binding:"required"`
			LoginType string `form:"login_type"`
		}

		var err error
		var url string

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode, url = h.service.OauthUrl(h.Context(c), req.OauthType, req.LoginType)
		}

		h.I18n.JSON(c, errCode, map[string]string{"url": url}, err)
	}
}

func (h handler) Token() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params system.AuthParam
		var token system.AccessToken
		var err error

		errCode := e.InvalidParams

		if err = c.ShouldBindJSON(&params); err == nil {
			token, errCode, err = h.service.Token(h.Context(c), &params)
		}

		h.I18n.JSON(c, errCode, token, err)
	}
}

// NewHandler creates and returns a new authentication Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewAuthService(
			appCtx.Redis["go-api"],
			appCtx.Logger,
			appCtx.MysqlDB["go-api"],
			appCtx.Notify,
		),
	}
}
