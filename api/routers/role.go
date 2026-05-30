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

func RoleRouter(rbacService *services.RBACService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "/batch", CreateRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodGet, "", ListRolesHandler(rbacService)),
		ginutils.WithHandler(http.MethodGet, "/:role_id", GetRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodPut, "/:role_id", UpdateRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/:role_id", DeleteRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/:role_id/permission-groups", AddPermissionGroupsToRoleHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/:role_id/permission-groups", RemovePermissionGroupsFromRoleHandler(rbacService)),
	)
}

// CreateRolesHandler godoc
// @Summary 批量创建角色
// @Tags role
// @Accept json
// @Produce json
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

// ListRolesHandler godoc
// @Summary 获取当前账号角色列表
// @Tags role
// @Produce json
// @Success 200 {object} ginutils.Ret[[]response.RoleResponse]
// @Router /roles [get]
func ListRolesHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		roles, err := rbacService.ListRoles(c.Request.Context())
		if response.HandleRouterError(c, err, "list roles failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToRoleListResponse(roles)))
	})
}

// GetRoleHandler godoc
// @Summary 获取角色详情
// @Tags role
// @Produce json
// @Param role_id path int true "角色ID"
// @Success 200 {object} ginutils.Ret[response.RoleResponse]
// @Router /roles/{role_id} [get]
func GetRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoleIDCommand) {
		role, err := rbacService.GetRole(c.Request.Context(), req.RoleID)
		if response.HandleRouterError(c, err, "get role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToRoleResponse(role)))
	})
}

// UpdateRoleHandler godoc
// @Summary 更新角色
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body requests.UpdateRoleCommand true "角色信息"
// @Success 200 {object} ginutils.Ret[response.RoleResponse]
// @Router /roles/{role_id} [put]
func UpdateRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateRoleCommand) {
		role, err := rbacService.UpdateRole(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToRoleResponse(role)))
	})
}

// DeleteRoleHandler godoc
// @Summary 删除角色
// @Tags role
// @Produce json
// @Param role_id path int true "角色ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /roles/{role_id} [delete]
func DeleteRoleHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RoleIDCommand) {
		err := rbacService.DeleteRole(c.Request.Context(), req.RoleID)
		if response.HandleRouterError(c, err, "delete role failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AddPermissionGroupsToRoleHandler godoc
// @Summary 给角色绑定权限组
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[any]
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
// @Summary 从角色移除权限组
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body requests.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[any]
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
