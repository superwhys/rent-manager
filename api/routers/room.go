package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "github.com/superwhys/rent-manager/api/response"
	appmeterreading "github.com/superwhys/rent-manager/internal/app/meterreading"
	apprentbill "github.com/superwhys/rent-manager/internal/app/rentbill"
	approom "github.com/superwhys/rent-manager/internal/app/room"
	"github.com/superwhys/rent-manager/internal/xerr"
)

// RoomRouter 注册房间路由。
func RoomRouter(roomService *approom.Service, meterReadingService *appmeterreading.Service, rentBillService *apprentbill.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/:room_id", GetRoomHandler(roomService)),
		ginutils.WithHandler(http.MethodPut, "/:room_id", UpdateRoomHandler(roomService)),
		ginutils.WithHandler(http.MethodDelete, "/:room_id", DeleteRoomHandler(roomService)),
		ginutils.WithHandler(http.MethodGet, "/:room_id/meter-readings", GetRoomMeterReadingsHandler(meterReadingService)),
		ginutils.WithHandler(http.MethodPut, "/:room_id/meter-readings", SaveRoomMeterReadingsHandler(meterReadingService)),
		ginutils.WithHandler(http.MethodGet, "/:room_id/rent-bills", ListRoomRentBillsHandler(rentBillService)),
		ginutils.WithHandler(http.MethodPost, "/:room_id/rent-bills/generate", GenerateRoomRentBillHandler(rentBillService)),
	)
}

// CreateRoomInUnitHandler godoc
// @Summary 在单元下创建房间
// @Tags room
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param request body room.CreateRoomInUnitCommand true "房间信息"
// @Success 200 {object} ginutils.Ret[room.RoomResponse]
// @Security BearerAuth
// @Router /units/{unit_id}/rooms [post]
func CreateRoomInUnitHandler(roomService *approom.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approom.CreateRoomInUnitCommand) {
		room, err := roomService.CreateRoomInUnit(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create room failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approom.ToRoomResponse(room)))
	})
}

// ListRoomsInUnitHandler godoc
// @Summary 获取单元房间列表
// @Tags room
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Param room_type_id query int false "房型ID"
// @Param is_active query bool false "是否可用"
// @Success 200 {object} ginutils.Ret[[]room.RoomResponse]
// @Security BearerAuth
// @Router /units/{unit_id}/rooms [get]
func ListRoomsInUnitHandler(roomService *approom.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approom.UnitRoomsQuery) {
		rooms, err := roomService.ListRoomsInUnit(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list rooms failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approom.ToRoomListResponse(rooms)))
	})
}

// GetRoomHandler godoc
// @Summary 获取房间详情
// @Tags room
// @Produce json
// @Param room_id path int true "房间ID"
// @Success 200 {object} ginutils.Ret[room.RoomResponse]
// @Security BearerAuth
// @Router /rooms/{room_id} [get]
func GetRoomHandler(roomService *approom.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approom.RoomIDQuery) {
		room, err := roomService.GetRoom(c.Request.Context(), req.RoomID)
		if apiresponse.HandleRouterError(c, err, "get room failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approom.ToRoomResponse(room)))
	})
}

// UpdateRoomHandler godoc
// @Summary 更新房间
// @Tags room
// @Accept json
// @Produce json
// @Param room_id path int true "房间ID"
// @Param request body room.UpdateRoomCommand true "房间信息"
// @Success 200 {object} ginutils.Ret[room.RoomResponse]
// @Security BearerAuth
// @Router /rooms/{room_id} [put]
func UpdateRoomHandler(roomService *approom.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approom.UpdateRoomCommand) {
		room, err := roomService.UpdateRoom(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "update room failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approom.ToRoomResponse(room)))
	})
}

// DeleteRoomHandler godoc
// @Summary 删除房间
// @Tags room
// @Produce json
// @Param room_id path int true "房间ID"
// @Success 200 {object} ginutils.Ret[any]
// @Security BearerAuth
// @Router /rooms/{room_id} [delete]
func DeleteRoomHandler(roomService *approom.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approom.RoomIDQuery) {
		err := roomService.DeleteRoom(c.Request.Context(), req.RoomID)
		if apiresponse.HandleRouterError(c, err, "delete room failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}
