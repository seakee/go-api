package user

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// Handler defines the interface for user-related HTTP handlers.
type Handler interface {
	i()
	Paginate() gin.HandlerFunc
	Detail() gin.HandlerFunc
	Create() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc

	Roles() gin.HandlerFunc
	UpdateRole() gin.HandlerFunc
	ResetPassword() gin.HandlerFunc
	DisableTfa() gin.HandlerFunc
	Passkeys() gin.HandlerFunc
	DeletePasskey() gin.HandlerFunc
	DeleteAllPasskeys() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.UserService
}

type info struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	TotpEnabled  bool      `json:"totp_enabled"`
	PasskeyCount int64     `json:"passkey_count"`
	UserName     string    `json:"user_name"`
	Status       int8      `json:"status"`
	Avatar       string    `json:"avatar"`
	CreatedAt    time.Time `json:"created_at"`
}

// params defines the user request parameters.
type params struct {
	ID       uint   `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	Password string `json:"password" binding:"required_without=ID"`
	Avatar   string `json:"avatar"`
}

func (h handler) Roles() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var data []uint
		var rid uint64

		userID := c.Query("user_id")
		errCode := e.InvalidParams

		ctx := h.Context(c)

		if userID != "" {
			errCode = e.ERROR
			rid, err = strconv.ParseUint(userID, 10, 64)
			if err != nil {
				h.I18n.JSON(c, e.InvalidParams, nil, err)
				return
			}
			data, err = h.service.Roles(ctx, uint(rid))
			if err == nil {
				errCode = e.SUCCESS
			}
		}

		h.I18n.JSON(c, errCode, data, err)
	}
}

func (h handler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID  uint   `json:"user_id"`
			RoleIDs []uint `json:"role_ids"`
		}

		ctx := h.Context(c)

		if err = c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
			h.I18n.JSON(c, e.InvalidParams, nil, err)
			return
		}

		errCode, err := h.service.UpdateRole(ctx, req.UserID, req.RoleIDs)
		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID       uint   `json:"user_id" binding:"required"`
			ReauthTicket string `json:"reauth_ticket" binding:"required"`
			Password     string `json:"password" binding:"required"`
		}

		ctx := h.Context(c)
		operatorUserID, _ := c.Get("user_id")

		errCode := e.InvalidParams
		if err = c.ShouldBindJSON(&req); err == nil {
			if req.UserID == 0 {
				h.I18n.JSON(c, e.InvalidParams, nil, nil)
				return
			}

			errCode, err = h.service.ResetPassword(ctx, operatorUserID.(uint), req.UserID, req.ReauthTicket, req.Password)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) DisableTfa() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID       uint   `json:"user_id" binding:"required"`
			ReauthTicket string `json:"reauth_ticket" binding:"required"`
		}

		ctx := h.Context(c)
		operatorUserID, _ := c.Get("user_id")

		errCode := e.InvalidParams
		if err = c.ShouldBindJSON(&req); err == nil {
			if req.UserID == 0 {
				h.I18n.JSON(c, e.InvalidParams, nil, nil)
				return
			}

			errCode, err = h.service.DisableTfa(ctx, operatorUserID.(uint), req.UserID, req.ReauthTicket)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) Passkeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID uint `form:"user_id" binding:"required"`
		}

		ctx := h.Context(c)
		errCode := e.InvalidParams
		if err = c.ShouldBindQuery(&req); err == nil {
			list, code, listErr := h.service.ListPasskeys(ctx, req.UserID)
			errCode = code
			err = listErr
			h.I18n.JSON(c, errCode, gin.H{"list": list}, err)
			return
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) DeletePasskey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID       uint   `json:"user_id" binding:"required"`
			ID           uint   `json:"id" binding:"required"`
			ReauthTicket string `json:"reauth_ticket" binding:"required"`
		}

		ctx := h.Context(c)
		operatorUserID, _ := c.Get("user_id")
		errCode := e.InvalidParams
		if err = c.ShouldBindJSON(&req); err == nil {
			errCode, err = h.service.DeletePasskey(ctx, operatorUserID.(uint), req.UserID, req.ID, req.ReauthTicket)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) DeleteAllPasskeys() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			UserID       uint   `json:"user_id" binding:"required"`
			ReauthTicket string `json:"reauth_ticket" binding:"required"`
		}

		ctx := h.Context(c)
		operatorUserID, _ := c.Get("user_id")
		errCode := e.InvalidParams
		if err = c.ShouldBindJSON(&req); err == nil {
			errCode, err = h.service.DeleteAllPasskeys(ctx, operatorUserID.(uint), req.UserID, req.ReauthTicket)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) i() {}

func (h handler) Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []systemModel.User
		var total int64
		var err error

		var req struct {
			UserName string `form:"user_name"`
			Status   int8   `form:"status"`
			Page     int    `form:"page"`
			PageSize int    `form:"page_size"`
			Email    string `form:"email"`
			Phone    string `form:"phone"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode = e.SUCCESS

			list, total, err = h.service.Paginate(ctx, &systemModel.User{
				UserName: req.UserName,
				Status:   req.Status,
				Email:    req.Email,
				Phone:    req.Phone,
			}, req.Page, req.PageSize)
			if err != nil {
				errCode = e.ERROR
			}
		}

		userIDs := make([]uint, 0, len(list))
		for i := range list {
			userIDs = append(userIDs, list[i].ID)
		}
		passkeyCounts, countErr := h.service.PasskeyCountByUserIDs(ctx, userIDs)
		if countErr != nil {
			errCode = e.ERROR
			err = countErr
		}

		userList := make([]info, len(list))
		for i := range list {
			userList[i] = info{
				ID:           list[i].ID,
				Email:        list[i].Email,
				Phone:        list[i].Phone,
				TotpEnabled:  list[i].TotpEnabled,
				PasskeyCount: passkeyCounts[list[i].ID],
				UserName:     list[i].UserName,
				Status:       list[i].Status,
				Avatar:       list[i].Avatar,
				CreatedAt:    list[i].CreatedAt,
			}
		}

		h.I18n.JSON(c, errCode, gin.H{
			"list":  userList,
			"total": total,
		}, err)
	}
}

func (h handler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var a *systemModel.User
		var data info

		var req struct {
			ID uint `form:"id" binding:"required"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams
		if err = c.ShouldBindQuery(&req); err == nil {
			a, errCode, err = h.service.Detail(ctx, req.ID)
			if err == nil {
				passkeyCount, countErr := h.service.PasskeyCount(ctx, req.ID)
				if countErr != nil {
					h.I18n.JSON(c, e.ERROR, nil, countErr)
					return
				}
				data = info{
					ID:           a.ID,
					Email:        a.Email,
					Phone:        a.Phone,
					TotpEnabled:  a.TotpEnabled,
					PasskeyCount: passkeyCount,
					UserName:     a.UserName,
					Status:       a.Status,
					Avatar:       a.Avatar,
					CreatedAt:    a.CreatedAt,
				}
			}
		}

		h.I18n.JSON(c, errCode, data, err)
	}
}

func (h handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var perm *systemModel.User

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			perm = &systemModel.User{
				UserName: req.UserName,
				Email:    req.Email,
				Phone:    req.Phone,
				Password: req.Password,
				Avatar:   req.Avatar,
				Status:   req.Status,
			}

			errCode, err = h.service.Create(ctx, perm)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		var req struct {
			ID uint `form:"id" binding:"required"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams
		if err = c.ShouldBindQuery(&req); err == nil {
			errCode, err = h.service.Delete(ctx, req.ID)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var a *systemModel.User

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			if req.ID != 0 {
				a = &systemModel.User{
					Model:    gorm.Model{ID: req.ID},
					UserName: req.UserName,
					Email:    req.Email,
					Phone:    req.Phone,
					Password: req.Password,
					Avatar:   req.Avatar,
					Status:   req.Status,
				}

				errCode, err = h.service.Update(ctx, a)
			}
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

// NewHandler creates and returns a new user Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewUserService(
			appCtx.Redis["go-api"],
			appCtx.Logger,
			appCtx.SqlDB["go-api"],
		),
	}
}
