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
}

type handler struct {
	controller.BaseController
	service system.UserService
}

type info struct {
	ID          uint      `json:"id"`
	Account     string    `json:"account"`
	FeishuID    string    `json:"feishu_id"`
	WechatID    string    `json:"wechat_id"`
	TotpEnabled bool      `json:"totp_enabled"`
	UserName    string    `json:"user_name"`
	Status      int8      `json:"status"`
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"created_at"`
}

// params defines the user request parameters.
type params struct {
	ID       uint   `json:"id"`
	UserName string `json:"user_name"`
	Account  string `json:"account"`
	Status   int8   `json:"status"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
	FeishuID string `json:"feishu_id"`
	WechatID string `json:"wechat_id"`
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

		if err = h.service.UpdateRole(ctx, req.UserID, req.RoleIDs); err != nil {
			h.I18n.JSON(c, e.ERROR, nil, err)
			return
		}

		h.I18n.JSON(c, e.SUCCESS, nil, nil)
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
			Account  string `form:"account"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode = e.SUCCESS

			list, total, err = h.service.Paginate(ctx, &systemModel.User{
				UserName: req.UserName,
				Status:   req.Status,
				Account:  req.Account,
			}, req.Page, req.PageSize)
			if err != nil {
				errCode = e.ERROR
			}
		}

		userList := make([]info, len(list))
		for i := range list {
			userList[i] = info{
				ID:          list[i].ID,
				Account:     list[i].Account,
				FeishuID:    list[i].FeishuId,
				WechatID:    list[i].WechatId,
				TotpEnabled: list[i].TotpEnabled,
				UserName:    list[i].UserName,
				Status:      list[i].Status,
				Avatar:      list[i].Avatar,
				CreatedAt:   list[i].CreatedAt,
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
				data = info{
					ID:          a.ID,
					Account:     a.Account,
					FeishuID:    a.FeishuId,
					WechatID:    a.WechatId,
					TotpEnabled: a.TotpEnabled,
					UserName:    a.UserName,
					Status:      a.Status,
					Avatar:      a.Avatar,
					CreatedAt:   a.CreatedAt,
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
				Account:  req.Account,
				Password: req.Password,
				Avatar:   req.Avatar,
				Status:   req.Status,
				FeishuId: req.FeishuID,
				WechatId: req.WechatID,
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
			errCode = e.ERROR
			err = h.service.Delete(ctx, req.ID)
			if err == nil {
				errCode = e.SUCCESS
			}
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
					Account:  req.Account,
					Password: req.Password,
					Avatar:   req.Avatar,
					Status:   req.Status,
					FeishuId: req.FeishuID,
					WechatId: req.WechatID,
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
