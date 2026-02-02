package record

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/service/system"
)

// Handler defines the interface for operation record-related HTTP handlers.
type Handler interface {
	i()

	Paginate() gin.HandlerFunc
	Interaction() gin.HandlerFunc
}

type handler struct {
	controller.BaseController
	service system.OperationRecordService
}

func (h handler) i() {}

func (h handler) Paginate() gin.HandlerFunc {
	return func(c *gin.Context) {
		errCode := e.ERROR
		var req struct {
			ID       string `form:"id"`
			Path     string `form:"path"`
			UserID   uint   `form:"user_id"`
			UserName string `form:"user_name"`
			Page     int    `form:"page"`
			Size     int    `form:"size"`
			IP       string `form:"ip"`
			Status   int    `form:"status"`
			Method   string `form:"method"`
		}
		if err := c.ShouldBind(&req); err != nil {
			h.I18n.JSON(c, e.InvalidParams, nil, err)
			return
		}

		r := &systemModel.OperationRecord{
			Path:     req.Path,
			UserID:   req.UserID,
			UserName: req.UserName,
			IP:       req.IP,
			Status:   req.Status,
			Method:   req.Method,
		}

		if req.ID != "" {
			_ = r.SetID(req.ID)
		}

		list, total, err := h.service.Paginate(
			h.Context(c),
			r,
			req.Page,
			req.Size,
		)
		if err == nil {
			errCode = e.SUCCESS
		}

		h.I18n.JSON(c, errCode, gin.H{
			"items": list,
			"total": total,
		}, err)
	}
}

func (h handler) Interaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID string `form:"id"`
		}

		ctx := h.Context(c)

		if err := c.ShouldBindQuery(&req); err != nil {
			h.I18n.JSON(c, e.InvalidParams, nil, err)
			return
		}

		data, err := h.service.Interaction(ctx, req.ID)
		if err != nil {
			h.I18n.JSON(c, e.ERROR, nil, err)
			return
		}

		h.I18n.JSON(c, e.SUCCESS, data, nil)
	}
}

// NewHandler creates and returns a new operation record Handler instance.
func NewHandler(appCtx *http.Context) Handler {
	return &handler{
		BaseController: controller.BaseController{
			AppCtx: appCtx,
			Logger: appCtx.Logger,
			Redis:  appCtx.Redis["go-api"],
			I18n:   appCtx.I18n,
		},
		service: system.NewOperationRecordService(
			appCtx.Redis["go-api"],
			appCtx.Logger,
			appCtx.SqlDB["go-api"],
		),
	}
}
