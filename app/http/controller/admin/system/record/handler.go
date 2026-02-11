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
	Detail() gin.HandlerFunc
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
			Path    string `form:"path"`
			UserID  uint   `form:"user_id"`
			Page    int    `form:"page"`
			Size    int    `form:"size"`
			IP      string `form:"ip"`
			Status  *int   `form:"status"`
			Method  string `form:"method"`
			TraceID string `form:"trace_id"`
		}
		if err := c.ShouldBind(&req); err != nil {
			h.I18n.JSON(c, e.InvalidParams, nil, err)
			return
		}

		r := &systemModel.OperationRecord{
			Path:    req.Path,
			UserID:  req.UserID,
			IP:      req.IP,
			Method:  req.Method,
			TraceID: req.TraceID,
		}

		if req.Status != nil {
			r = r.Where("status = ?", *req.Status)
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

func (h handler) Detail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID string `form:"id"`
		}

		ctx := h.Context(c)

		if err := c.ShouldBindQuery(&req); err != nil {
			h.I18n.JSON(c, e.InvalidParams, nil, err)
			return
		}

		if req.ID == "" {
			h.I18n.JSON(c, e.InvalidParams, nil, nil)
			return
		}

		data, err := h.service.Detail(ctx, req.ID)
		if err != nil {
			h.I18n.JSON(c, e.ERROR, nil, err)
			return
		}

		if data == nil {
			h.I18n.JSON(c, e.SUCCESS, nil, nil)
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
