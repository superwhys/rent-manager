package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	apprbac "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rbac"
	approom "git.superwhys.top/rent-manager/rent-manager-server/internal/app/room"
	appunit "git.superwhys.top/rent-manager/rent-manager-server/internal/app/unit"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// UnitRouter 注册单元及单元下资源路由。
func UnitRouter(unitService *appunit.Service, roomService *approom.Service, rbacService *apprbac.Service) ginutils.Option {
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
// @Param request body unit.CreateUnitCommand true "单元信息"
// @Success 200 {object} ginutils.Ret[unit.UnitResponse]
// @Router /units [post]
func CreateUnitHandler(unitService *appunit.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appunit.CreateUnitCommand) {
		unit, err := unitService.CreateUnit(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create unit failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appunit.ToUnitResponse(unit)))
	})
}

// ListUnitsHandler godoc
// @Summary 获取单元列表
// @Tags unit
// @Produce json
// @Param scope query string false "列表范围: all/owned/granted，默认 all"
// @Success 200 {object} ginutils.Ret[[]unit.UnitResponse]
// @Router /units [get]
func ListUnitsHandler(unitService *appunit.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appunit.ListUnitsQuery) {
		summaries, err := unitService.ListUnits(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list units failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appunit.ToUnitListResponse(summaries)))
	})
}

// GetUnitHandler godoc
// @Summary 获取单元详情
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Success 200 {object} ginutils.Ret[unit.UnitResponse]
// @Router /units/{unit_id} [get]
func GetUnitHandler(unitService *appunit.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appunit.UnitIDQuery) {
		unit, err := unitService.GetUnit(c.Request.Context(), req.UnitID)
		if apiresponse.HandleRouterError(c, err, "get unit failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appunit.ToUnitResponse(unit)))
	})
}

// UpdateUnitHandler godoc
// @Summary 更新单元
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param request body unit.UpdateUnitCommand true "单元信息"
// @Success 200 {object} ginutils.Ret[unit.UnitResponse]
// @Router /units/{unit_id} [put]
func UpdateUnitHandler(unitService *appunit.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appunit.UpdateUnitCommand) {
		unit, err := unitService.UpdateUnit(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "update unit failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appunit.ToUnitResponse(unit)))
	})
}

// DeleteUnitHandler godoc
// @Summary 删除单元
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /units/{unit_id} [delete]
func DeleteUnitHandler(unitService *appunit.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appunit.UnitIDQuery) {
		err := unitService.DeleteUnit(c.Request.Context(), req.UnitID)
		if apiresponse.HandleRouterError(c, err, "delete unit failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}

// ListUnitGrantsHandler godoc
// @Summary 获取单元授权列表
// @Tags unit
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id query int false "被授权账号ID"
// @Param status query string false "授权状态"
// @Success 200 {object} ginutils.Ret[[]rbac.UnitGrantResponse]
// @Router /units/{unit_id}/grants [get]
func ListUnitGrantsHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.UnitGrantsQuery) {
		grants, err := rbacService.ListUnitGrants(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "list unit grants failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToUnitGrantListResponse(grants)))
	})
}

// GrantUnitRolesHandler godoc
// @Summary 授权账号管理单元角色
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id path int true "被授权账号ID"
// @Param request body rbac.UnitAccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[[]rbac.UnitGrantResponse]
// @Router /units/{unit_id}/accounts/{account_id}/roles [post]
func GrantUnitRolesHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.UnitAccountRolesCommand) {
		grants, err := rbacService.GrantUnitRoles(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "grant unit roles failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToUnitGrantListResponse(grants)))
	})
}

// RevokeUnitRolesHandler godoc
// @Summary 撤销账号单元角色授权
// @Tags unit
// @Accept json
// @Produce json
// @Param unit_id path int true "单元ID"
// @Param account_id path int true "被授权账号ID"
// @Param request body rbac.UnitAccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /units/{unit_id}/accounts/{account_id}/roles [delete]
func RevokeUnitRolesHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.UnitAccountRolesCommand) {
		err := rbacService.RevokeUnitRoles(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "revoke unit roles failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}
