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

func RoomRouter(roomService *services.RoomService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/:room_id", GetRoomHandler(roomService)),
		ginutils.WithHandler(http.MethodPut, "/:room_id", UpdateRoomHandler(roomService)),
		ginutils.WithHandler(http.MethodDelete, "/:room_id", DeleteRoomHandler(roomService)),
	)
}

// CreateRoomInUnitHandler godoc
// @Summary 在单元下创建房间
// @Tags room
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param request body requests.CreateRoomInUnitCommand true "房间信息"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /units/{unit_id}/rooms [post]
func CreateRoomInUnitHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRoomInUnitCommand) {
		room, err := roomService.CreateRoomInUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
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
// @Param is_active query bool false "是否启用"
// @Success 200 {object} ginutils.Ret[[]response.RoomResponse]
// @Router /units/{unit_id}/rooms [get]
func ListRoomsInUnitHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitRoomsQuery) {
		rooms, err := roomService.ListRoomsInUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list rooms failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomListResponseRet(rooms))
	})
}

// GetRoomHandler godoc
// @Summary 获取房间详情
// @Tags room
// @Produce json
// @Param room_id path int true "房间ID"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /rooms/{room_id} [get]
func GetRoomHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomIdQuery) {
		room, err := roomService.GetRoom(c.Request.Context(), req.RoomID)
		if response.HandleRouterError(c, err, "get room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}

// UpdateRoomHandler godoc
// @Summary 更新房间
// @Tags room
// @Accept json
// @Produce json
// @Param room_id path int true "房间ID"
// @Param request body requests.UpdateRoomCommand true "房间信息"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /rooms/{room_id} [put]
func UpdateRoomHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateRoomCommand) {
		room, err := roomService.UpdateRoom(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}

// DeleteRoomHandler godoc
// @Summary 删除房间
// @Tags room
// @Produce json
// @Param room_id path int true "房间ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /rooms/{room_id} [delete]
func DeleteRoomHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomIdQuery) {
		err := roomService.DeleteRoom(c.Request.Context(), req.RoomID)
		if response.HandleRouterError(c, err, "delete room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
