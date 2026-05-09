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

func AuthGroupRouter(authService *services.AuthService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/auth"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/wx-login", WechatLoginHandler(authService)),
			ginutils.WithHandler(http.MethodPost, "/refresh", AccountRefreshHandler(authService)),
		),
	)
}

// WechatLoginHandler 处理微信登录
// @Summary 微信登录
// @Description 微信登录
// @Tags Account
// @Accept json
// @Produce json
// @Param request body requests.WechatLoginReq true "微信登录请求体"
// @Success 200 {object} ginutils.Ret[response.TokenResponse]
// @Router /auth/wx-login [post]
func WechatLoginHandler(authService *services.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.WechatLoginReq) {
		token, err := authService.WechatLogin(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "auth wechat login handler error", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(token))
	})
}

// AccountRefreshHandler 刷新 access token
// @Summary 刷新 access token
// @Description 使用 refresh token 刷新 access token
// @Tags Account
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenReq true "刷新 Token 请求体"
// @Success 200 {object} ginutils.Ret[response.TokenResponse]
// @Router /auth/refresh [post]
func AccountRefreshHandler(authService *services.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RefreshTokenReq) {
		token, err := authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
		if response.HandleRouterError(c, err, "auth refresh handler error", errcode.ErrInvalidToken) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(token))
	})
}
