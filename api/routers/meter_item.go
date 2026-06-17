package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/app/meteritem"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// MeterItemRouter 注册抄表项路由。
func MeterItemRouter(meterItemService *appmeteritem.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "", CreateMeterItemHandler(meterItemService)),
		ginutils.WithHandler(http.MethodGet, "", ListMeterItemsHandler(meterItemService)),
		ginutils.WithHandler(http.MethodGet, "/:meter_item_id", GetMeterItemHandler(meterItemService)),
		ginutils.WithHandler(http.MethodPut, "/:meter_item_id", UpdateMeterItemHandler(meterItemService)),
	)
}

// CreateMeterItemHandler godoc
// @Summary 创建抄表项
// @Tags meter_item
// @Accept json
// @Produce json
// @Param request body meteritem.CreateMeterItemCommand true "抄表项信息"
// @Success 200 {object} ginutils.Ret[meteritem.MeterItemResponse]
// @Security BearerAuth
// @Router /meter-items [post]
func CreateMeterItemHandler(meterItemService *appmeteritem.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeteritem.CreateMeterItemCommand) {
		item, err := meterItemService.CreateMeterItem(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create meter item failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeteritem.ToMeterItemResponse(item)))
	})
}

// ListMeterItemsHandler godoc
// @Summary 获取抄表项列表
// @Tags meter_item
// @Produce json
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Param code query string false "抄表项编码"
// @Param is_system query bool false "是否系统内置"
// @Param calc_type query string false "计费类型"
// @Success 200 {object} ginutils.Ret[[]meteritem.MeterItemResponse]
// @Security BearerAuth
// @Router /meter-items [get]
func ListMeterItemsHandler(meterItemService *appmeteritem.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeteritem.ListMeterItemsQuery) {
		items, err := meterItemService.ListMeterItems(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list meter items failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeteritem.ToMeterItemListResponse(items)))
	})
}

// GetMeterItemHandler godoc
// @Summary 获取抄表项详情
// @Tags meter_item
// @Produce json
// @Param meter_item_id path int true "抄表项ID"
// @Success 200 {object} ginutils.Ret[meteritem.MeterItemResponse]
// @Security BearerAuth
// @Router /meter-items/{meter_item_id} [get]
func GetMeterItemHandler(meterItemService *appmeteritem.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeteritem.MeterItemIDQuery) {
		item, err := meterItemService.GetMeterItem(c.Request.Context(), req.MeterItemID)
		if apiresponse.HandleRouterError(c, err, "get meter item failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeteritem.ToMeterItemResponse(item)))
	})
}

// UpdateMeterItemHandler godoc
// @Summary 更新抄表项
// @Tags meter_item
// @Accept json
// @Produce json
// @Param meter_item_id path int true "抄表项ID"
// @Param request body meteritem.UpdateMeterItemCommand true "抄表项信息"
// @Success 200 {object} ginutils.Ret[meteritem.MeterItemResponse]
// @Security BearerAuth
// @Router /meter-items/{meter_item_id} [put]
func UpdateMeterItemHandler(meterItemService *appmeteritem.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeteritem.UpdateMeterItemCommand) {
		item, err := meterItemService.UpdateMeterItem(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "update meter item failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeteritem.ToMeterItemResponse(item)))
	})
}
