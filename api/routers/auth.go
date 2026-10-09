package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "github.com/superwhys/rent-manager/api/response"
	appauth "github.com/superwhys/rent-manager/internal/app/auth"
	"github.com/superwhys/rent-manager/internal/xerr"
)

// AuthRouter 注册认证路由。
func AuthRouter(authService *appauth.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "/wechat-login", WechatLoginHandler(authService)),
		ginutils.WithHandler(http.MethodPost, "/refresh-token", RefreshTokenHandler(authService)),
	)
}

// WechatLoginHandler godoc
// @Summary 微信登录
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.WechatLoginReq true "微信登录参数"
// @Success 200 {object} ginutils.Ret[auth.TokenResponse]
// @Router /auth/wechat-login [post]
func WechatLoginHandler(authService *appauth.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.WechatLoginReq) {
		token, err := authService.WechatLogin(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "wechat login failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appauth.ToTokenResponse(token)))
	})
}

// RefreshTokenHandler godoc
// @Summary 刷新访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenReq true "刷新令牌参数"
// @Success 200 {object} ginutils.Ret[auth.TokenResponse]
// @Router /auth/refresh-token [post]
func RefreshTokenHandler(authService *appauth.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.RefreshTokenReq) {
		token, err := authService.RefreshToken(c.Request.Context(), req)
		if apiresponse.HandleRouterError(c, err, "refresh token failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appauth.ToTokenResponse(token)))
	})
}
