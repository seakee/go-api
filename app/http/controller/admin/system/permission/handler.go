package permission

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
	"gorm.io/gorm"
)

// Handler defines the interface for permission-related HTTP handlers.
type Handler interface {
	i()

	Available() gin.HandlerFunc
	List() gin.HandlerFunc
	Paginate() gin.HandlerFunc
	Detail() gin.HandlerFunc
	Create() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.PermissionService
}

// params defines the permission request parameters.
type params struct {
	ID          uint   `json:"id"`
	Name        string `json:"name" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Group       string `json:"group" binding:"required"`
}

func (h handler) i() {}

func (h handler) Available() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.Context(c)

		errCode := e.ERROR

		available, err := h.service.Available(ctx)
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, available, err)
	}
}

func (h handler) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var list map[string][]systemModel.Permission

		var req struct {
			Type string `form:"type" binding:"required"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams
		if err = c.ShouldBindQuery(&req); err == nil {
			errCode = e.ERROR
			list, err = h.service.ListInGroup(ctx, req.Type)
			if err == nil {
				errCode = e.SUCCESS
			}
		}

		h.I18n.JSON(c, errCode, list, err)
	}
}

func (h handler) Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []systemModel.Permission
		var total int64
		var err error

		var req struct {
			Name     string `form:"name"`
			Group    string `form:"group"`
			Method   string `form:"method"`
			Page     int    `form:"page"`
			PageSize int    `form:"page_size"`
			Type     string `form:"type" binding:"required"`
		}

		ctx := h.Context(c)

		errCode := e.InvalidParams

		if err = c.ShouldBind(&req); err == nil {
			errCode = e.SUCCESS

			list, total, err = h.service.Paginate(ctx, &systemModel.Permission{
				Name:   req.Name,
				Group:  req.Group,
				Type:   req.Type,
				Method: req.Method,
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
		var perm *systemModel.Permission

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
		var perm *systemModel.Permission

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			perm = &systemModel.Permission{
				Name:        req.Name,
				Method:      req.Method,
				Path:        req.Path,
				Description: req.Description,
				Type:        req.Type,
				Group:       req.Group,
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
		var perm *systemModel.Permission

		ctx := h.Context(c)

		errCode := e.InvalidParams
		req := params{}
		if err = c.ShouldBindJSON(&req); err == nil {
			if req.ID != 0 {
				perm = &systemModel.Permission{
					Model:       gorm.Model{ID: req.ID},
					Name:        req.Name,
					Method:      req.Method,
					Path:        req.Path,
					Description: req.Description,
					Type:        req.Type,
					Group:       req.Group,
				}

				errCode, err = h.service.Update(ctx, perm)
			}
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

// NewHandler creates and returns a new permission Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewPermissionService(
			appCtx.Redis["go-api"],
			appCtx.Logger,
			appCtx.MysqlDB["go-api"],
			appCtx.Engine.Routes,
		),
	}
}
