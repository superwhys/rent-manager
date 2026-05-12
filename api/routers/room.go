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

func RoomGroupRouter(roomService *services.RoomService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/rooms"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/", RoomCreateHandler(roomService)),
			ginutils.WithHandler(http.MethodGet, "/", RoomListHandler(roomService)),
			ginutils.WithHandler(http.MethodGet, "/:room_id", RoomGetHandler(roomService)),
			ginutils.WithHandler(http.MethodPut, "/:room_id", RoomUpdateHandler(roomService)),
			ginutils.WithHandler(http.MethodDelete, "/:room_id", RoomDeleteHandler(roomService)),
		),
	)
}

// RoomCreateHandler 处理创建房间请求。
// @Summary 创建房间
// @Description 创建房间
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateRoomCommand true "创建房间请求体"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /rooms [post]
func RoomCreateHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRoomCommand) {
		room, err := roomService.CreateRoom(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}

// RoomListHandler 处理房间列表查询请求。
// @Summary 查询房间列表
// @Description 查询房间列表
// @Tags Room
// @Produce json
// @Security BearerAuth
// @Param request query requests.ListRoomsQuery true "获取房间列表请求参数"
// @Success 200 {object} ginutils.Ret[[]response.RoomResponse]
// @Router /rooms [get]
func RoomListHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListRoomsQuery) {
		rooms, err := roomService.ListRooms(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list rooms failed", errcode.ErrSysInternal) {
			return
		}

		c.JSON(http.StatusOK, response.RoomListResponseRet(rooms))
	})
}

// RoomGetHandler 处理单个房间查询请求。
// @Summary 查询房间
// @Description 查询单个房间
// @Tags Room
// @Produce json
// @Security BearerAuth
// @Param room_id path int true "房间 ID"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /rooms/{room_id} [get]
func RoomGetHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomIdQuery) {
		room, err := roomService.GetRoom(c.Request.Context(), req.OrganizationID, req.RoomID)
		if response.HandleRouterError(c, err, "get room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}

// RoomUpdateHandler 处理更新房间请求。
// @Summary 更新房间
// @Description 更新房间
// @Tags Room
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param room_id path int true "房间 ID"
// @Param request body requests.UpdateRoomCommand true "更新房间请求体"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /rooms/{room_id} [put]
func RoomUpdateHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateRoomCommand) {
		room, err := roomService.UpdateRoom(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}

// RoomDeleteHandler 处理删除房间请求。
// @Summary 删除房间
// @Description 删除房间
// @Tags Room
// @Produce json
// @Security BearerAuth
// @Param room_id path int true "房间 ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /rooms/{room_id} [delete]
func RoomDeleteHandler(roomService *services.RoomService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomIdQuery) {
		err := roomService.DeleteRoom(c.Request.Context(), req.OrganizationID, req.RoomID)
		if response.HandleRouterError(c, err, "delete room failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
