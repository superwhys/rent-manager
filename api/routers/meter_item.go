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

func MeterItemRouter(meterItemService *services.MeterItemService) ginutils.Option {
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
// @Param request body requests.CreateMeterItemCommand true "抄表项信息"
// @Success 200 {object} ginutils.Ret[response.MeterItemResponse]
// @Router /meter-items [post]
func CreateMeterItemHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateMeterItemCommand) {
		item, err := meterItemService.CreateMeterItem(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create meter item failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemResponseRet(item))
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
// @Success 200 {object} ginutils.Ret[[]response.MeterItemResponse]
// @Router /meter-items [get]
func ListMeterItemsHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListMeterItemsQuery) {
		items, err := meterItemService.ListMeterItems(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list meter items failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemListResponseRet(items))
	})
}

// GetMeterItemHandler godoc
// @Summary 获取抄表项详情
// @Tags meter_item
// @Produce json
// @Param meter_item_id path int true "抄表项ID"
// @Success 200 {object} ginutils.Ret[response.MeterItemResponse]
// @Router /meter-items/{meter_item_id} [get]
func GetMeterItemHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.MeterItemIDQuery) {
		item, err := meterItemService.GetMeterItem(c.Request.Context(), req.MeterItemID)
		if response.HandleRouterError(c, err, "get meter item failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemResponseRet(item))
	})
}

// UpdateMeterItemHandler godoc
// @Summary 更新抄表项
// @Tags meter_item
// @Accept json
// @Produce json
// @Param meter_item_id path int true "抄表项ID"
// @Param request body requests.UpdateMeterItemCommand true "抄表项信息"
// @Success 200 {object} ginutils.Ret[response.MeterItemResponse]
// @Router /meter-items/{meter_item_id} [put]
func UpdateMeterItemHandler(meterItemService *services.MeterItemService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateMeterItemCommand) {
		item, err := meterItemService.UpdateMeterItem(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update meter item failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.MeterItemResponseRet(item))
	})
}
