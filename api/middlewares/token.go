package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/jwtutils"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

func TokenVerifyMiddleware(jwtConfig *config.JwtConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrUnauthorized))
			ctx.Abort()
			return
		}

		// Support "Bearer <token>" format
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := jwtutils.ParseToken(tokenStr, []byte(jwtConfig.JwtSecret))
		if err != nil {
			logging.Errorw(ctx.Request.Context(), "verify token failed", logging.Err(err))

			// Handle token expiration specifically
			if errors.Is(err, jwt.ErrTokenExpired) {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrTokenExpired))
			} else if ec, ok := errcode.AsErrcode(err); ok {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(ec))
			} else {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrInvalidToken))
			}
			ctx.Abort()
			return
		}

		logging.Debugw(ctx.Request.Context(), "token claims", logging.Any("claims", claims))

		ctx.Set(string(jwtutils.TokenContextKey), claims)

		reqCtx := jwtutils.SetTokenClaimsToContext(ctx.Request.Context(), claims)
		reqCtx = logging.With(reqCtx, "UserID", claims.UserID)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}
