package system

import (
	"github.com/gin-gonic/gin"
	"github.com/seakee/go-api/app/http"
	"github.com/seakee/go-api/app/http/controller/admin/system/record"
)

func registerRecordRoutes(api *gin.RouterGroup, ctx *http.Context) {
	handler := record.NewHandler(ctx)

	api.GET("paginate", handler.Paginate())
	api.GET("interaction", handler.Interaction())
}
