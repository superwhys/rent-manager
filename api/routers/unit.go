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

func UnitGroupRouter(unitService *services.UnitService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/units"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/", UnitCreateHandler(unitService)),
			ginutils.WithHandler(http.MethodGet, "/", UnitListHandler(unitService)),
			ginutils.WithHandler(http.MethodGet, "/:unit_id", UnitGetHandler(unitService)),
			ginutils.WithHandler(http.MethodPut, "/:unit_id", UnitUpdateHandler(unitService)),
			ginutils.WithHandler(http.MethodDelete, "/:unit_id", UnitDeleteHandler(unitService)),
			ginutils.WithHandler(http.MethodPost, "/:unit_id/rooms", UnitCreateRoomHandler(unitService)),
		),
	)
}

// UnitCreateHandler 处理创建单元请求。
// @Summary 创建单元
// @Description 创建单元并设置默认抄表项
// @Tags Unit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateUnitCommand true "创建单元请求体"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units [post]
func UnitCreateHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateUnitCommand) {
		unit, err := unitService.CreateUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// UnitListHandler 处理单元列表查询请求。
// @Summary 查询单元列表
// @Description 查询当前账号的单元列表
// @Tags Unit
// @Produce json
// @Security BearerAuth
// @Param request query requests.ListUnitsQuery true "获取单元列表请求参数"
// @Success 200 {object} ginutils.Ret[[]response.UnitResponse]
// @Router /units [get]
func UnitListHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListUnitsQuery) {
		units, err := unitService.ListUnits(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list units failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitListResponseRet(units))
	})
}

// UnitGetHandler 处理单元详情查询请求。
// @Summary 查询单元详情
// @Description 查询当前账号的单元详情
// @Tags Unit
// @Produce json
// @Security BearerAuth
// @Param unit_id path int true "单元 ID"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units/{unit_id} [get]
func UnitGetHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitIDQuery) {
		unit, err := unitService.GetUnit(c.Request.Context(), req.UnitID)
		if response.HandleRouterError(c, err, "get unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// UnitUpdateHandler 处理更新单元请求。
// @Summary 更新单元
// @Description 更新单元并整体替换默认抄表项
// @Tags Unit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unit_id path int true "单元 ID"
// @Param request body requests.UpdateUnitCommand true "更新单元请求体"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units/{unit_id} [put]
func UnitUpdateHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateUnitCommand) {
		unit, err := unitService.UpdateUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// UnitDeleteHandler 处理删除单元请求。
// @Summary 删除单元
// @Description 删除没有房间的单元
// @Tags Unit
// @Produce json
// @Security BearerAuth
// @Param unit_id path int true "单元 ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /units/{unit_id} [delete]
func UnitDeleteHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitIDQuery) {
		err := unitService.DeleteUnit(c.Request.Context(), req.UnitID)
		if response.HandleRouterError(c, err, "delete unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// UnitCreateRoomHandler 处理在单元下创建房间请求。
// @Summary 在单元下创建房间
// @Description 在指定单元下创建房间
// @Tags Unit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unit_id path int true "单元 ID"
// @Param request body requests.CreateRoomInUnitCommand true "创建房间请求体"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /units/{unit_id}/rooms [post]
func UnitCreateRoomHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRoomInUnitCommand) {
		room, err := unitService.CreateRoomInUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create room in unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}
