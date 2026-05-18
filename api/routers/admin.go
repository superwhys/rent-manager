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

func AdminRouter(rbacService *services.RBACService) ginutils.Option {
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
// @Param request body requests.CreatePermissionGroupsCommand true "权限组列表"
// @Success 200 {object} ginutils.Ret[[]response.PermissionGroupResponse]
// @Router /admin/permission-groups/batch [post]
func CreatePermissionGroupsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreatePermissionGroupsCommand) {
		groups, err := rbacService.CreatePermissionGroups(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create permission groups failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToPermissionGroupListResponse(groups)))
	})
}

// ListPermissionGroupsHandler godoc
// @Summary 获取权限组列表
// @Tags admin
// @Produce json
// @Success 200 {object} ginutils.Ret[[]response.PermissionGroupResponse]
// @Router /admin/permission-groups [get]
func ListPermissionGroupsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		groups, err := rbacService.ListPermissionGroups(c.Request.Context())
		if response.HandleRouterError(c, err, "list permission groups failed", errcode.ErrSysInternal) {
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

// ListPermissionsHandler godoc
// @Summary 获取权限列表
// @Tags admin
// @Produce json
// @Success 200 {object} ginutils.Ret[[]response.PermissionResponse]
// @Router /admin/permissions [get]
func ListPermissionsHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		permissions, err := rbacService.ListPermissions(c.Request.Context())
		if response.HandleRouterError(c, err, "list permissions failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToPermissionListResponse(permissions)))
	})
}

// AddPermissionsToPermissionGroupHandler godoc
// @Summary 给权限组绑定权限
// @Tags admin
// @Accept json
// @Produce json
// @Param permission_group_id path int true "权限组ID"
// @Param request body requests.PermissionGroupPermissionsCommand true "权限ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /admin/permission-groups/{permission_group_id}/permissions [post]
func AddPermissionsToPermissionGroupHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.PermissionGroupPermissionsCommand) {
		err := rbacService.AddPermissionsToPermissionGroup(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "add permissions to permission group failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RemovePermissionsFromPermissionGroupHandler godoc
// @Summary 从权限组移除权限
// @Tags admin
// @Accept json
// @Produce json
// @Param permission_group_id path int true "权限组ID"
// @Param request body requests.PermissionGroupPermissionsCommand true "权限ID列表"
// @Success 200 {object} ginutils.Ret[any]
// @Router /admin/permission-groups/{permission_group_id}/permissions [delete]
func RemovePermissionsFromPermissionGroupHandler(rbacService *services.RBACService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.PermissionGroupPermissionsCommand) {
		err := rbacService.RemovePermissionsFromPermissionGroup(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "remove permissions from permission group failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
