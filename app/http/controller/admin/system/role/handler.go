package role

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
	"gorm.io/gorm"
	"strconv"
)

// Handler defines the interface for role-related HTTP handlers.
type Handler interface {
	i()
	List() gin.HandlerFunc
	Paginate() gin.HandlerFunc
	Detail() gin.HandlerFunc
	Create() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
	Permissions() gin.HandlerFunc
	UpdatePermission() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.RoleService
}

// params defines the role request parameters.
type params struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h handler) Permissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var data []uint
		var rid uint64

		roleId := c.Query("role_id")
		errCode := e.InvalidParams

		ctx := h.Context(c)

		if roleId != "" {
			errCode = e.ERROR
			rid, err = strconv.ParseUint(roleId, 10, 64)
			data, err = h.service.PermissionList(ctx, uint(rid))
			if err == nil {
				errCode = e.SUCCESS
			}
		}

		h.I18n.JSON(c, errCode, data, err)
	}
}

func (h handler) UpdatePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var req struct {
			RoleID        uint   `json:"role_id" binding:"required"`
			PermissionIDs []uint `json:"permission_ids"`
		}

		ctx := h.Context(c)
		errCode := e.InvalidParams
		if err = c.ShouldBindJSON(&req); err != nil || req.RoleID == 0 {
			h.I18n.JSON(c, errCode, nil, err)
			return
		}

		errCode, err = h.service.UpdatePermission(ctx, req.RoleID, req.PermissionIDs)

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) i() {}

func (h handler) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var list []gin.H

		ctx := h.Context(c)

		errCode := e.ERROR
		list, err = h.service.List(ctx)
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, list, err)
	}
}

func (h handler) Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []systemModel.Role
		var total int64
		var err error

		var req struct {
			Name     string `form:"name"`
			Page     int    `form:"page"`
			PageSize int    `form:"page_size"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode = e.SUCCESS

			list, total, err = h.service.Paginate(ctx, &systemModel.Role{
				Name: req.Name,
			}, req.Page, req.PageSize)
			if err != nil {
				errCode = e.ERROR
			}
		}

		h.I18n.JSON(c, errCode, gin.H{
			"list":  list,
			"total": total,
		}, err)
	}
}

func (h handler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var perm *systemModel.Role

		var req struct {
			ID uint `form:"id" binding:"required"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams
		if err = c.ShouldBindQuery(&req); err == nil {
			perm, errCode, err = h.service.Detail(ctx, req.ID)
		}

		h.I18n.JSON(c, errCode, perm, err)
	}
}

func (h handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var perm *systemModel.Role

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			perm = &systemModel.Role{
				Name:        req.Name,
				Description: req.Description,
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
		var perm *systemModel.Role

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			if req.ID != 0 {
				perm = &systemModel.Role{
					Model:       gorm.Model{ID: req.ID},
					Name:        req.Name,
					Description: req.Description,
				}

				errCode, err = h.service.Update(ctx, perm)
			}
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

// NewHandler creates and returns a new role Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewRoleService(
			appCtx.Redis["go-api"],
			appCtx.Logger,
			appCtx.SqlDB["go-api"],
		),
	}
}
