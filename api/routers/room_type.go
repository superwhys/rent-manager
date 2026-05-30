package routers

import (
	"net/http"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/app/services"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/errcode"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/response"
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

func RoomTypeRouter(roomTypeService *services.RoomTypeService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "", CreateRoomTypeHandler(roomTypeService)),
		ginutils.WithHandler(http.MethodGet, "", ListRoomTypesHandler(roomTypeService)),
		ginutils.WithHandler(http.MethodGet, "/:room_type_id", GetRoomTypeHandler(roomTypeService)),
		ginutils.WithHandler(http.MethodPut, "/:room_type_id", UpdateRoomTypeHandler(roomTypeService)),
		ginutils.WithHandler(http.MethodDelete, "/:room_type_id", DeleteRoomTypeHandler(roomTypeService)),
	)
}

// CreateRoomTypeHandler godoc
// @Summary 创建房型
// @Tags room_type
// @Accept json
// @Produce json
// @Param request body requests.CreateRoomTypeCommand true "房型信息"
// @Success 200 {object} ginutils.Ret[response.RoomTypeResponse]
// @Router /room-types [post]
func CreateRoomTypeHandler(roomTypeService *services.RoomTypeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRoomTypeCommand) {
		roomType, err := roomTypeService.CreateRoomType(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create room type failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomTypeResponseRet(roomType))
	})
}

// ListRoomTypesHandler godoc
// @Summary 获取房型列表
// @Tags room_type
// @Produce json
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Param name query string false "房型名称"
// @Success 200 {object} ginutils.Ret[[]response.RoomTypeResponse]
// @Router /room-types [get]
func ListRoomTypesHandler(roomTypeService *services.RoomTypeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListRoomTypesQuery) {
		roomTypes, err := roomTypeService.ListRoomTypes(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list room types failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomTypeListResponseRet(roomTypes))
	})
}

// GetRoomTypeHandler godoc
// @Summary 获取房型详情
// @Tags room_type
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Success 200 {object} ginutils.Ret[response.RoomTypeResponse]
// @Router /room-types/{room_type_id} [get]
func GetRoomTypeHandler(roomTypeService *services.RoomTypeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomTypeIDQuery) {
		roomType, err := roomTypeService.GetRoomType(c.Request.Context(), req.RoomTypeID)
		if response.HandleRouterError(c, err, "get room type failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomTypeResponseRet(roomType))
	})
}

// UpdateRoomTypeHandler godoc
// @Summary 更新房型
// @Tags room_type
// @Accept json
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Param request body requests.UpdateRoomTypeCommand true "房型信息"
// @Success 200 {object} ginutils.Ret[response.RoomTypeResponse]
// @Router /room-types/{room_type_id} [put]
func UpdateRoomTypeHandler(roomTypeService *services.RoomTypeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateRoomTypeCommand) {
		roomType, err := roomTypeService.UpdateRoomType(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update room type failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomTypeResponseRet(roomType))
	})
}

// DeleteRoomTypeHandler godoc
// @Summary 删除房型
// @Tags room_type
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room-types/{room_type_id} [delete]
func DeleteRoomTypeHandler(roomTypeService *services.RoomTypeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoomTypeIDQuery) {
		err := roomTypeService.DeleteRoomType(c.Request.Context(), req.RoomTypeID)
		if response.HandleRouterError(c, err, "delete room type failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
