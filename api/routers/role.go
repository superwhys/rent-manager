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

func RoleGroupRouter(rbacService *services.RBACService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "/roles/batch", CreateRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/roles/:role_id/permission-groups", AddPermissionGroupsToRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/roles/:role_id/permission-groups", RemovePermissionGroupsFromRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/units/:unit_id/accounts/:account_id/roles", AssignRolesToUnitAccountHandler(rbacService)),
	)
}

// CreateRolesHandler godoc
// @Summary 批量创建账号角色
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateRolesCommand true "角色列表"
// @Success 200 {object} ginutils.Ret[[]response.RoleResponse]
// @Router /roles/batch [post]
func CreateRolesHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRolesCommand) {
		roles, err := rbacService.CreateRoles(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create roles failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToRoleListResponse(roles)))
	})
}

// AddPermissionGroupsToRoleHandler godoc
// @Summary 给账号角色绑定权限组
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /roles/{role_id}/permission-groups [post]
func AddPermissionGroupsToRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionGroupsCommand) {
		err := rbacService.AddPermissionGroupsToRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "add permission groups to role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RemovePermissionGroupsFromRoleHandler godoc
// @Summary 从账号角色移除权限组
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /roles/{role_id}/permission-groups [delete]
func RemovePermissionGroupsFromRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionGroupsCommand) {
		err := rbacService.RemovePermissionGroupsFromRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "remove permission groups from role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AssignRolesToUnitAccountHandler godoc
// @Summary 给账号授予单元管理角色
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unit_id path int true "单元ID"
// @Param account_id path int true "账号ID"
// @Param request body requests.UnitAccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /units/{unit_id}/accounts/{account_id}/roles [post]
func AssignRolesToUnitAccountHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UnitAccountRolesCommand) {
		err := rbacService.AssignRolesToUnitAccount(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "assign roles to unit account failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
