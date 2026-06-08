package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	approomtype "git.superwhys.top/rent-manager/rent-manager-server/internal/app/roomtype"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// RoomTypeRouter 注册房型路由。
func RoomTypeRouter(roomTypeService *approomtype.Service) ginutils.Option {
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
// @Param request body roomtype.CreateRoomTypeCommand true "房型信息"
// @Success 200 {object} ginutils.Ret[roomtype.RoomTypeResponse]
// @Router /room-types [post]
func CreateRoomTypeHandler(roomTypeService *approomtype.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approomtype.CreateRoomTypeCommand) {
		roomType, err := roomTypeService.CreateRoomType(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create room type failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approomtype.ToRoomTypeResponse(roomType)))
	})
}

// ListRoomTypesHandler godoc
// @Summary 获取房型列表
// @Tags room_type
// @Produce json
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Param name query string false "房型名称"
// @Success 200 {object} ginutils.Ret[[]roomtype.RoomTypeResponse]
// @Router /room-types [get]
func ListRoomTypesHandler(roomTypeService *approomtype.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approomtype.ListRoomTypesQuery) {
		roomTypes, err := roomTypeService.ListRoomTypes(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list room types failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approomtype.ToRoomTypeListResponse(roomTypes)))
	})
}

// GetRoomTypeHandler godoc
// @Summary 获取房型详情
// @Tags room_type
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Success 200 {object} ginutils.Ret[roomtype.RoomTypeResponse]
// @Router /room-types/{room_type_id} [get]
func GetRoomTypeHandler(roomTypeService *approomtype.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approomtype.RoomTypeIDQuery) {
		roomType, err := roomTypeService.GetRoomType(c.Request.Context(), req.RoomTypeID)
		if apiresponse.HandleRouterError(c, err, "get room type failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approomtype.ToRoomTypeResponse(roomType)))
	})
}

// UpdateRoomTypeHandler godoc
// @Summary 更新房型
// @Tags room_type
// @Accept json
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Param request body roomtype.UpdateRoomTypeCommand true "房型信息"
// @Success 200 {object} ginutils.Ret[roomtype.RoomTypeResponse]
// @Router /room-types/{room_type_id} [put]
func UpdateRoomTypeHandler(roomTypeService *approomtype.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approomtype.UpdateRoomTypeCommand) {
		roomType, err := roomTypeService.UpdateRoomType(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "update room type failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(approomtype.ToRoomTypeResponse(roomType)))
	})
}

// DeleteRoomTypeHandler godoc
// @Summary 删除房型
// @Tags room_type
// @Produce json
// @Param room_type_id path int true "房型ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room-types/{room_type_id} [delete]
func DeleteRoomTypeHandler(roomTypeService *approomtype.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *approomtype.RoomTypeIDQuery) {
		err := roomTypeService.DeleteRoomType(c.Request.Context(), req.RoomTypeID)
		if apiresponse.HandleRouterError(c, err, "delete room type failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}
