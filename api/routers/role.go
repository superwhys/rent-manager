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
		ginutils.WithPrefix("/organizations/:organization_id"),
		ginutils.WithHandler(http.MethodPost, "/roles/batch", CreateRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/roles/:role_id/permission-groups", AddPermissionGroupsToRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/roles/:role_id/permission-groups", RemovePermissionGroupsFromRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/accounts/:account_id/roles", AssignRolesToAccountHandler(rbacService)),
	)
}

// CreateRolesHandler godoc
// @Summary 批量创建组织角色
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path int true "组织ID"
// @Param request body requests.CreateRolesCommand true "角色列表"
// @Success 200 {object} ginutils.Ret[[]response.RoleResponse]
// @Router /organizations/{organization_id}/roles/batch [post]
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
// @Summary 给组织角色绑定权限组
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path int true "组织ID"
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /organizations/{organization_id}/roles/{role_id}/permission-groups [post]
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
// @Summary 从组织角色移除权限组
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path int true "组织ID"
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /organizations/{organization_id}/roles/{role_id}/permission-groups [delete]
func RemovePermissionGroupsFromRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionGroupsCommand) {
		err := rbacService.RemovePermissionGroupsFromRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "remove permission groups from role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AssignRolesToAccountHandler godoc
// @Summary 给组织内账号分配角色
// @Tags role
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param organization_id path int true "组织ID"
// @Param account_id path int true "账号ID"
// @Param request body requests.AccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[string]
// @Router /organizations/{organization_id}/accounts/{account_id}/roles [post]
func AssignRolesToAccountHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.AccountRolesCommand) {
		err := rbacService.AssignRolesToAccount(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "assign roles to account failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
