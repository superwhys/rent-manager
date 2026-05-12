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
		ginutils.WithHandler(http.MethodPost, "/permission-groups/batch", CreatePermissionGroupsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/permissions/batch", CreatePermissionsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/permission-groups/:permission_group_id/permissions", AddPermissionsToPermissionGroupHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/permission-groups/:permission_group_id/permissions", RemovePermissionsFromPermissionGroupHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/roles/:role_id/permission-groups", AddPermissionGroupsToRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/roles/:role_id/permission-groups", RemovePermissionGroupsFromRoleHandler(rbacService)),
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
// @Success 200 {object} ginutils.Ret[[]response.RoleResponse]
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

func CreatePermissionGroupsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreatePermissionGroupsCommand) {
		groups, err := rbacService.CreatePermissionGroups(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create permission groups failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToPermissionGroupListResponse(groups)))
	})
}

// CreatePermissionsHandler godoc
// @Summary 批量创建权限
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreatePermissionsCommand true "权限列表"
// @Success 200 {object} ginutils.Ret[[]response.PermissionResponse]
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

func AddPermissionsToPermissionGroupHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.PermissionGroupPermissionsCommand) {
		err := rbacService.AddPermissionsToPermissionGroup(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "add permissions to permission group failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

func RemovePermissionsFromPermissionGroupHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.PermissionGroupPermissionsCommand) {
		err := rbacService.RemovePermissionsFromPermissionGroup(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "remove permissions from permission group failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

func AddPermissionGroupsToRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RolePermissionGroupsCommand) {
		err := rbacService.AddPermissionGroupsToRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "add permission groups to role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

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
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param account_id path int true "账号ID"
// @Param request body requests.AccountRolesCommand true "角色ID列表"
// @Success 200 {object} ginutils.Ret[string]
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
