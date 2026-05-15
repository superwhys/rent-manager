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
		ginutils.WithHandler(http.MethodPost, "/permission-groups/batch", CreatePermissionGroupsHandler(rbacService)),
	)
}

// CreatePermissionGroupsHandler godoc
// @Summary 批量创建权限组
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
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
