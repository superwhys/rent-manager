package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

func MeterItemGroupRouter(meterItemService *services.MeterItemService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/meter-items"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/", MeterItemCreateHandler(meterItemService)),
			ginutils.WithHandler(http.MethodGet, "/", MeterItemListHandler(meterItemService)),
			ginutils.WithHandler(http.MethodPut, "/:meter_item_id", MeterItemUpdateHandler(meterItemService)),
		),
	)
}

// MeterItemCreateHandler 处理创建抄表项请求。
// @Summary 创建抄表项
// @Description 为当前账号创建抄表项
// @Tags MeterItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateMeterItemCommand true "创建抄表项请求体"
// @Success 200 {object} ginutils.Ret[response.MeterItemResponse]
// @Router /meter-items [post]
func MeterItemCreateHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateMeterItemCommand) {
		item, err := meterItemService.CreateMeterItem(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create meter item failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemResponseRet(item))
	})
}

// MeterItemListHandler 处理抄表项列表查询请求。
// @Summary 查询抄表项列表
// @Description 查询当前账号的抄表项列表
// @Tags MeterItem
// @Produce json
// @Security BearerAuth
// @Param request query requests.ListMeterItemsQuery true "获取抄表项列表请求参数"
// @Success 200 {object} ginutils.Ret[[]response.MeterItemResponse]
// @Router /meter-items [get]
func MeterItemListHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListMeterItemsQuery) {
		items, err := meterItemService.ListMeterItems(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list meter items failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemListResponseRet(items))
	})
}

// MeterItemUpdateHandler 处理更新抄表项请求。
// @Summary 更新抄表项
// @Description 更新当前账号的抄表项
// @Tags MeterItem
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param meter_item_id path int true "抄表项 ID"
// @Param request body requests.UpdateMeterItemCommand true "更新抄表项请求体"
// @Success 200 {object} ginutils.Ret[response.MeterItemResponse]
// @Router /meter-items/{meter_item_id} [put]
func MeterItemUpdateHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateMeterItemCommand) {
		item, err := meterItemService.UpdateMeterItem(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update meter item failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemResponseRet(item))
	})
}
