package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	apprbac "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rbac"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// RoleRouter 注册角色管理路由。
func RoleRouter(rbacService *apprbac.Service) ginutils.Option {
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
// @Param request body rbac.CreateRolesCommand true "角色列表"
// @Success 200 {object} ginutils.Ret[[]rbac.RoleResponse]
// @Security BearerAuth
// @Router /roles/batch [post]
func CreateRolesHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.CreateRolesCommand) {
		roles, err := rbacService.CreateRoles(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create roles failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToRoleListResponse(roles)))
	})
}

// ListRolesHandler godoc
// @Summary 获取当前账号角色列表
// @Tags role
// @Produce json
// @Success 200 {object} ginutils.Ret[[]rbac.RoleResponse]
// @Security BearerAuth
// @Router /roles [get]
func ListRolesHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		roles, err := rbacService.ListRoles(c.Request.Context())
		if apiresponse.HandleRouterError(c, err, "list roles failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToRoleListResponse(roles)))
	})
}

// GetRoleHandler godoc
// @Summary 获取角色详情
// @Tags role
// @Produce json
// @Param role_id path int true "角色ID"
// @Success 200 {object} ginutils.Ret[rbac.RoleResponse]
// @Security BearerAuth
// @Router /roles/{role_id} [get]
func GetRoleHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.RoleIDCommand) {
		role, err := rbacService.GetRole(c.Request.Context(), req.RoleID)
		if apiresponse.HandleRouterError(c, err, "get role failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToRoleResponse(role)))
	})
}

// UpdateRoleHandler godoc
// @Summary 更新角色
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body rbac.UpdateRoleCommand true "角色信息"
// @Success 200 {object} ginutils.Ret[rbac.RoleResponse]
// @Security BearerAuth
// @Router /roles/{role_id} [put]
func UpdateRoleHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.UpdateRoleCommand) {
		role, err := rbacService.UpdateRole(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "update role failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToRoleResponse(role)))
	})
}

// DeleteRoleHandler godoc
// @Summary 删除角色
// @Tags role
// @Produce json
// @Param role_id path int true "角色ID"
// @Success 200 {object} ginutils.Ret[any]
// @Security BearerAuth
// @Router /roles/{role_id} [delete]
func DeleteRoleHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.RoleIDCommand) {
		err := rbacService.DeleteRole(c.Request.Context(), req.RoleID)
		if apiresponse.HandleRouterError(c, err, "delete role failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}

// AddPermissionGroupsToRoleHandler godoc
// @Summary 给角色绑定权限组
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body rbac.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Security BearerAuth
// @Router /roles/{role_id}/permission-groups [post]
func AddPermissionGroupsToRoleHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.RolePermissionGroupsCommand) {
		err := rbacService.AddPermissionGroupsToRole(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "add permission groups to role failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}

// RemovePermissionGroupsFromRoleHandler godoc
// @Summary 从角色移除权限组
// @Tags role
// @Accept json
// @Produce json
// @Param role_id path int true "角色ID"
// @Param request body rbac.RolePermissionGroupsCommand true "权限组ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Security BearerAuth
// @Router /roles/{role_id}/permission-groups [delete]
func RemovePermissionGroupsFromRoleHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.RolePermissionGroupsCommand) {
		err := rbacService.RemovePermissionGroupsFromRole(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "remove permission groups from role failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}
