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

func UnitRouter(unitService *services.UnitService, roomService *services.RoomService, rbacService *services.RBACService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "", CreateUnitHandler(unitService)),
		ginutils.WithHandler(http.MethodGet, "", ListUnitsHandler(unitService)),
		ginutils.WithHandler(http.MethodGet, "/:unit_id", GetUnitHandler(unitService)),
		ginutils.WithHandler(http.MethodPut, "/:unit_id", UpdateUnitHandler(unitService)),
		ginutils.WithHandler(http.MethodDelete, "/:unit_id", DeleteUnitHandler(unitService)),
		ginutils.WithHandler(http.MethodPost, "/:unit_id/rooms", CreateRoomInUnitHandler(roomService)),
		ginutils.WithHandler(http.MethodGet, "/:unit_id/rooms", ListRoomsInUnitHandler(roomService)),
		ginutils.WithHandler(http.MethodGet, "/:unit_id/grants", ListUnitGrantsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/:unit_id/accounts/:account_id/roles", GrantUnitRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/:unit_id/accounts/:account_id/roles", RevokeUnitRolesHandler(rbacService)),
	)
}

// CreateUnitHandler godoc
// @Summary 创建单元
// @Tags unit
// @Accept json
// @Produce json
// @Param request body requests.CreateUnitCommand true "单元信息"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units [post]
func CreateUnitHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateUnitCommand) {
		unit, err := unitService.CreateUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// ListUnitsHandler godoc
// @Summary 获取单元列表
// @Tags unit
// @Produce json
// @Param limit query int false "分页数量"
// @Param offset query int false "分页偏移"
// @Param name query string false "单元名称"
// @Param city query string false "城市"
// @Param province query string false "省份"
// @Param scope query string false "列表范围: all/owned/granted，默认 all"
// @Success 200 {object} ginutils.Ret[[]response.UnitResponse]
// @Router /units [get]
func ListUnitsHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListUnitsQuery) {
		units, err := unitService.ListUnits(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list units failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitListResponseRet(units))
	})
}

// GetUnitHandler godoc
// @Summary 获取单元详情
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units/{unit_id} [get]
func GetUnitHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitIDQuery) {
		unit, err := unitService.GetUnit(c.Request.Context(), req.UnitID)
		if response.HandleRouterError(c, err, "get unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// UpdateUnitHandler godoc
// @Summary 更新单元
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param request body requests.UpdateUnitCommand true "单元信息"
// @Success 200 {object} ginutils.Ret[response.UnitResponse]
// @Router /units/{unit_id} [put]
func UpdateUnitHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateUnitCommand) {
		unit, err := unitService.UpdateUnit(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.UnitResponseRet(unit))
	})
}

// DeleteUnitHandler godoc
// @Summary 删除单元
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /units/{unit_id} [delete]
func DeleteUnitHandler(unitService *services.UnitService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitIDQuery) {
		err := unitService.DeleteUnit(c.Request.Context(), req.UnitID)
		if response.HandleRouterError(c, err, "delete unit failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// ListUnitGrantsHandler godoc
// @Summary 获取单元授权列表
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id query int false "被授权账号ID"
// @Param status query string false "授权状态"
// @Success 200 {object} ginutils.Ret[[]response.UnitGrantResponse]
// @Router /units/{unit_id}/grants [get]
func ListUnitGrantsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitGrantsQuery) {
		grants, err := rbacService.ListUnitGrants(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list unit grants failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToUnitGrantListResponse(grants)))
	})
}

// GrantUnitRolesHandler godoc
// @Summary 授权账号管理单元角色
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id path int true "被授权账号ID"
// @Param request body requests.UnitAccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[[]response.UnitGrantResponse]
// @Router /units/{unit_id}/accounts/{account_id}/roles [post]
func GrantUnitRolesHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitAccountRolesCommand) {
		grants, err := rbacService.GrantUnitRoles(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "grant unit roles failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToUnitGrantListResponse(grants)))
	})
}

// RevokeUnitRolesHandler godoc
// @Summary 撤销账号单元角色授权
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id path int true "被授权账号ID"
// @Param request body requests.UnitAccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /units/{unit_id}/accounts/{account_id}/roles [delete]
func RevokeUnitRolesHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitAccountRolesCommand) {
		err := rbacService.RevokeUnitRoles(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "revoke unit roles failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
