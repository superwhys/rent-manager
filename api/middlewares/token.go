package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"

	"git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	appauth "git.superwhys.top/rent-manager/rent-manager-server/internal/app/auth"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// TokenVerifyMiddleware 校验访问令牌并写入当前账号上下文。
func TokenVerifyMiddleware(tokenManager appauth.TokenManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(xerr.ErrUnauthorized))
			ctx.Abort()
			return
		}

		// Support "Bearer <token>" format
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := tokenManager.ParseToken(tokenStr)
		if err != nil {
			logging.Errorw(ctx.Request.Context(), "verify token failed", logging.Err(err))

			if ec, ok := xerr.AsErrcode(err); ok {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(ec))
			} else {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(xerr.ErrInvalidToken))
			}
			ctx.Abort()
			return
		}

		logging.Debugw(ctx.Request.Context(), "token claims", logging.Any("claims", claims))

		reqCtx := appcore.SetAccountIDToContext(ctx.Request.Context(), claims.UserID)
		reqCtx = logging.With(reqCtx, "UserID", claims.UserID)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}
