package routers

import (
	"net/http"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	apprbac "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rbac"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

// AdminRouter 注册全局权限管理路由。
func AdminRouter(rbacService *apprbac.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "/permission-groups/batch", CreatePermissionGroupsHandler(rbacService)),
		ginutils.WithHandler(http.MethodGet, "/permission-groups", ListPermissionGroupsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/permissions/batch", CreatePermissionsHandler(rbacService)),
		ginutils.WithHandler(http.MethodGet, "/permissions", ListPermissionsHandler(rbacService)),
		ginutils.WithHandler(http.MethodPost, "/permission-groups/:permission_group_id/permissions", AddPermissionsToPermissionGroupHandler(rbacService)),
		ginutils.WithHandler(http.MethodDelete, "/permission-groups/:permission_group_id/permissions", RemovePermissionsFromPermissionGroupHandler(rbacService)),
	)
}

// CreatePermissionGroupsHandler godoc
// @Summary 批量创建权限组
// @Tags admin
// @Accept json
// @Produce json
// @Param request body rbac.CreatePermissionGroupsCommand true "权限组列表"
// @Success 200 {object} ginutils.Ret[[]rbac.PermissionGroupResponse]
// @Router /admin/permission-groups/batch [post]
func CreatePermissionGroupsHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.CreatePermissionGroupsCommand) {
		groups, err := rbacService.CreatePermissionGroups(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create permission groups failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToPermissionGroupListResponse(groups)))
	})
}

// ListPermissionGroupsHandler godoc
// @Summary 获取权限组列表
// @Tags admin
// @Produce json
// @Success 200 {object} ginutils.Ret[[]rbac.PermissionGroupResponse]
// @Router /admin/permission-groups [get]
func ListPermissionGroupsHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		groups, err := rbacService.ListPermissionGroups(c.Request.Context())
		if apiresponse.HandleRouterError(c, err, "list permission groups failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToPermissionGroupListResponse(groups)))
	})
}

// CreatePermissionsHandler godoc
// @Summary 批量创建权限
// @Tags admin
// @Accept json
// @Produce json
// @Param request body rbac.CreatePermissionsCommand true "权限列表"
// @Success 200 {object} ginutils.Ret[[]rbac.PermissionResponse]
// @Router /admin/permissions/batch [post]
func CreatePermissionsHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.CreatePermissionsCommand) {
		permissions, err := rbacService.CreatePermissions(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "create permissions failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToPermissionListResponse(permissions)))
	})
}

// ListPermissionsHandler godoc
// @Summary 获取权限列表
// @Tags admin
// @Produce json
// @Success 200 {object} ginutils.Ret[[]rbac.PermissionResponse]
// @Router /admin/permissions [get]
func ListPermissionsHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		permissions, err := rbacService.ListPermissions(c.Request.Context())
		if apiresponse.HandleRouterError(c, err, "list permissions failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apprbac.ToPermissionListResponse(permissions)))
	})
}

// AddPermissionsToPermissionGroupHandler godoc
// @Summary 给权限组绑定权限
// @Tags admin
// @Accept json
// @Produce json
// @Param permission_group_id path int true "权限组ID"
// @Param request body rbac.PermissionGroupPermissionsCommand true "权限ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /admin/permission-groups/{permission_group_id}/permissions [post]
func AddPermissionsToPermissionGroupHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.PermissionGroupPermissionsCommand) {
		err := rbacService.AddPermissionsToPermissionGroup(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "add permissions to permission group failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}

// RemovePermissionsFromPermissionGroupHandler godoc
// @Summary 从权限组移除权限
// @Tags admin
// @Accept json
// @Produce json
// @Param permission_group_id path int true "权限组ID"
// @Param request body rbac.PermissionGroupPermissionsCommand true "权限ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /admin/permission-groups/{permission_group_id}/permissions [delete]
func RemovePermissionsFromPermissionGroupHandler(rbacService *apprbac.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *apprbac.PermissionGroupPermissionsCommand) {
		err := rbacService.RemovePermissionsFromPermissionGroup(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "remove permissions from permission group failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	})
}
