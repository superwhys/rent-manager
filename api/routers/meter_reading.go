package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appmeterreading "git.superwhys.top/rent-manager/rent-manager-server/internal/app/meterreading"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// GetRoomMeterReadingsHandler godoc
// @Summary 获取房间月度抄表状态
// @Tags meter_reading
// @Produce json
// @Param room_id path int true "房间ID"
// @Param period_month query string true "抄表月份 YYYY-MM"
// @Success 200 {object} ginutils.Ret[meterreading.RoomMonthReadingsResponse]
// @Router /rooms/{room_id}/meter-readings [get]
func GetRoomMeterReadingsHandler(meterReadingService *appmeterreading.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeterreading.RoomMeterReadingsQuery) {
		roomMonth, err := meterReadingService.GetRoomMonthReadings(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "get room meter readings failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeterreading.ToRoomMonthReadingsResponse(roomMonth)))
	})
}

// SaveRoomMeterReadingsHandler godoc
// @Summary 保存房间月度抄表
// @Tags meter_reading
// @Accept json
// @Produce json
// @Param room_id path int true "房间ID"
// @Param request body meterreading.SaveRoomMonthCommand true "房间月度抄表"
// @Success 200 {object} ginutils.Ret[meterreading.RoomMonthReadingsResponse]
// @Router /rooms/{room_id}/meter-readings [put]
func SaveRoomMeterReadingsHandler(meterReadingService *appmeterreading.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeterreading.SaveRoomMonthCommand) {
		roomMonth, err := meterReadingService.SaveRoomMonthReadings(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "save room meter readings failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeterreading.ToRoomMonthReadingsResponse(roomMonth)))
	})
}

// GetUnitMeterReadingsHandler godoc
// @Summary 获取单元月度抄表汇总
// @Tags meter_reading
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param period_month query string true "抄表月份 YYYY-MM"
// @Success 200 {object} ginutils.Ret[meterreading.UnitMonthReadingsResponse]
// @Router /units/{unit_id}/meter-readings [get]
func GetUnitMeterReadingsHandler(meterReadingService *appmeterreading.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appmeterreading.UnitMeterReadingsQuery) {
		unitMonth, err := meterReadingService.GetUnitMonthReadings(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "get unit meter readings failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appmeterreading.ToUnitMonthReadingsResponse(unitMonth)))
	})
}
