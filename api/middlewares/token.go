package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/logging"

	"github.com/superwhys/rent-manager/api/response"
	appcore "github.com/superwhys/rent-manager/internal/app"
	"github.com/superwhys/rent-manager/internal/xerr"
)

type SessionAuthenticator interface {
	AuthenticateSession(context.Context, string) (*authkit.AuthenticatedSession, error)
}

const sessionContextKey = "authkit.session"

// BearerToken 提取 Authorization 中的会话令牌。
func BearerToken(ctx *gin.Context) string {
	parts := strings.Fields(ctx.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// AuthenticatedSessionFromContext 返回中间件已验证的 authkit 会话。
func AuthenticatedSessionFromContext(ctx *gin.Context) *authkit.AuthenticatedSession {
	value, _ := ctx.Get(sessionContextKey)
	session, _ := value.(*authkit.AuthenticatedSession)
	return session
}

// TokenVerifyMiddleware 使用 authkit 校验会话并写入当前账号上下文。
func TokenVerifyMiddleware(authService SessionAuthenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := BearerToken(ctx)
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(xerr.ErrUnauthorized))
			return
		}
		session, err := authService.AuthenticateSession(ctx.Request.Context(), token)
		if err != nil {
			if errors.Is(err, authkit.ErrUnauthorized) || errors.Is(err, authkit.ErrBlacklisted) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(xerr.ErrInvalidToken))
			} else {
				logging.Errorw(ctx.Request.Context(), "authenticate session failed", logging.Err(err))
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponseWithCode(xerr.ErrSysInternal))
			}
			return
		}
		// 本项目未提供代登录授权，不能把代登录会话当作目标账号的普通会话。
		if session == nil || session.Account.ID == "" || session.Actor != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(xerr.ErrInvalidToken))
			return
		}
		ctx.Set(sessionContextKey, session)
		reqCtx := appcore.SetAccountIDToContext(ctx.Request.Context(), session.Account.ID)
		reqCtx = logging.With(reqCtx, "AccountID", session.Account.ID)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}
