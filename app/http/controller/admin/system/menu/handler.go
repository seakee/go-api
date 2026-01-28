package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
	"strconv"
)

// Handler defines the interface for menu-related HTTP handlers.
type Handler interface {
	i()

	List() gin.HandlerFunc
	Detail() gin.HandlerFunc
	Create() gin.HandlerFunc
	Delete() gin.HandlerFunc
	Update() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.MenuService
}

func (h handler) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := h.Context(c)

		errCode := e.ERROR

		list, err := h.service.List(ctx)
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, gin.H{"items": list}, err)
	}
}

func (h handler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var menu *systemModel.Menu

		ctx := h.Context(c)

		errCode := e.InvalidParams
		id := c.Query("id")
		if id != "" {
			var uintID uint64

			errCode = e.ERROR
			uintID, err = strconv.ParseUint(id, 10, 64)
			if err == nil {
				menu, errCode, err = h.service.Detail(ctx, uint(uintID))
			}
		}

		h.I18n.JSON(c, errCode, menu, err)
	}
}

func (h handler) Create() gin.HandlerFunc {
	type createMenuParams struct {
		ParentID uint   `json:"parent_id"`
		Name     string `json:"name" binding:"required"`
		Icon     string `json:"icon"`
		Sort     *int   `json:"sort" binding:"required"` // Use int pointer to prevent required validation error when sort is 0
		Path     string `json:"path" binding:"required"`
	}

	return func(c *gin.Context) {
		var err error

		ctx := h.Context(c)

		errCode := e.InvalidParams

		menu := &systemModel.Menu{}
		params := &createMenuParams{}
		if err = c.ShouldBindJSON(&params); err == nil {
			menu.ParentId = params.ParentID
			menu.Name = params.Name
			menu.Icon = params.Icon
			menu.Sort = *params.Sort
			menu.Path = params.Path

			errCode, err = h.service.Create(ctx, menu)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		ctx := h.Context(c)

		errCode := e.InvalidParams
		id := c.Query("id")
		if id != "" {
			var uintID uint64

			errCode = e.ERROR
			uintID, err = strconv.ParseUint(id, 10, 64)
			if err == nil {
				errCode, err = h.service.Delete(ctx, uint(uintID))
			}
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) Update() gin.HandlerFunc {
	type updateMenuParams struct {
		ID   uint   `json:"id"  binding:"required"`
		Name string `json:"name"`
		Sort int    `json:"sort"`
		Icon string `json:"icon"`
		Path string `json:"path"`
	}

	return func(c *gin.Context) {
		var err error

		ctx := h.Context(c)

		errCode := e.InvalidParams

		menu := &systemModel.Menu{}
		params := &updateMenuParams{}
		if err = c.ShouldBindJSON(&params); err == nil {
			menu.ID = params.ID
			menu.Name = params.Name
			menu.Sort = params.Sort
			menu.Icon = params.Icon
			menu.Path = params.Path

			errCode, err = h.service.Update(ctx, menu)
		}

		h.I18n.JSON(c, errCode, nil, err)
	}
}

func (h handler) i() {}

// NewHandler creates and returns a new menu Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewMenuService(appCtx.Redis["go-api"], appCtx.Logger, appCtx.MysqlDB["go-api"]),
	}
}
