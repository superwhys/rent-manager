package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/jwtutils"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, accountID int, apiPath string) (bool, error)
	HasAccountRoles(ctx context.Context) (bool, error)
}

func PermissionMiddleware(checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwtutils.TokenClaimsFromContext(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrUnauthorized))
			return
		}
		if claims.TokenType != jwtutils.TokenTypeAccess {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrInvalidToken))
			return
		}

		apiPath := c.FullPath()
		if apiPath == "" {
			apiPath = c.Request.URL.Path
		}

		if isBootstrapAdminPath(apiPath) {
			hasAccountRoles, err := checker.HasAccountRoles(c.Request.Context())
			if err != nil {
				logging.Errorf("check account roles failed: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponseWithCode(errcode.ErrSysInternal))
				return
			}
			if !hasAccountRoles {
				c.Next()
				return
			}
		}

		hasPermission, err := checker.HasPermission(c.Request.Context(), claims.UserID, apiPath)
		if err != nil {
			logging.Errorf("check permission failed: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponseWithCode(errcode.ErrSysInternal))
			return
		}
		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponseWithCode(errcode.ErrForbidden))
			return
		}

		c.Next()
	}
}

func isBootstrapAdminPath(apiPath string) bool {
	switch apiPath {
	case "/admin/roles/batch", "/admin/permissions/batch", "/admin/accounts/:account_id/roles":
		return true
	default:
		return false
	}
}
