package middlewares

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/jwtutils"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, accountID int, organizationID int, apiPath string) (bool, error)
}

type SystemPermissionChecker interface {
	HasSystemPermission(ctx context.Context, accountID int) (bool, error)
}

func SystemAdminMiddleware(checker SystemPermissionChecker) gin.HandlerFunc {
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

		hasPermission, err := checker.HasSystemPermission(c.Request.Context(), claims.UserID)
		if err != nil {
			logging.Errorf("check system permission failed: %v", err)
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

		organizationID, ok := organizationIDFromRequest(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponseWithCode(errcode.ErrForbidden))
			return
		}

		apiPath := c.FullPath()
		if apiPath == "" {
			apiPath = c.Request.URL.Path
		}

		hasPermission, err := checker.HasPermission(c.Request.Context(), claims.UserID, organizationID, apiPath)
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

func organizationIDFromRequest(c *gin.Context) (int, bool) {
	values := []string{
		c.Param("organization_id"),
		c.Query("organization_id"),
		c.GetHeader("X-Organization-ID"),
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		id, err := strconv.Atoi(value)
		if err == nil && id > 0 {
			return id, true
		}
	}
	return 0, false
}
