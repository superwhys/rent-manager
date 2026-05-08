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

func AdminGroupRouter(rbacService *services.RBACService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/admin"),
		ginutils.WithHandler(http.MethodPost, "/roles/batch", CreateRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/permissions/batch", CreatePermissionsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/roles/:role_id/permissions", AddPermissionsToRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/roles/:role_id/permissions", RemovePermissionsFromRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/accounts/:account_id/roles", AssignRolesToAccountHandler(rbacService)),
	)
}

// CreateRolesHandler godoc
// @Summary 批量创建角色
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateRolesCommand true "角色列表"
// @Success 200 {object} response.Ret[[]response.RoleResponse]
// @Router /admin/roles/batch [post]
func CreateRolesHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRolesCommand) {
		roles, err := rbacService.CreateRoles(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create roles failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToRoleListResponse(roles)))
	})
}

// CreatePermissionsHandler godoc
// @Summary 批量创建权限
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreatePermissionsCommand true "权限列表"
// @Success 200 {object} response.Ret[[]response.PermissionResponse]
// @Router /admin/permissions/batch [post]
func CreatePermissionsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreatePermissionsCommand) {
		permissions, err := rbacService.CreatePermissions(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create permissions failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToPermissionListResponse(permissions)))
	})
}

// AddPermissionsToRoleHandler godoc
// @Summary 给角色添加权限
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionsCommand true "权限ID列表"
// @Success 200 {object} response.Ret[string]
// @Router /admin/roles/{role_id}/permissions [post]
func AddPermissionsToRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionsCommand) {
		err := rbacService.AddPermissionsToRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "add permissions to role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RemovePermissionsFromRoleHandler godoc
// @Summary 删除角色的部分权限
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionsCommand true "权限ID列表"
// @Success 200 {object} response.Ret[string]
// @Router /admin/roles/{role_id}/permissions [delete]
func RemovePermissionsFromRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionsCommand) {
		err := rbacService.RemovePermissionsFromRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "remove permissions from role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AssignRolesToAccountHandler godoc
// @Summary 给账号分配角色
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param account_id path int true "账号ID"
// @Param request body requests.AccountRolesCommand true "角色ID列表"
// @Success 200 {object} response.Ret[string]
// @Router /admin/accounts/{account_id}/roles [post]
func AssignRolesToAccountHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.AccountRolesCommand) {
		err := rbacService.AssignRolesToAccount(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "assign roles to account failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
