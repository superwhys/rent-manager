package routers

import (
	"net/http"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/app/services"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/errcode"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/response"
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

func AuthRouter(authService *services.AuthService) ginutils.Option {
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
// @Param request body requests.WechatLoginReq true "微信登录参数"
// @Success 200 {object} ginutils.Ret[response.TokenResponse]
// @Router /auth/wechat-login [post]
func WechatLoginHandler(authService *services.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.WechatLoginReq) {
		token, err := authService.WechatLogin(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "wechat login failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ToTokenResponseRet(token))
	})
}

// RefreshTokenHandler godoc
// @Summary 刷新访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenReq true "刷新令牌参数"
// @Success 200 {object} ginutils.Ret[response.TokenResponse]
// @Router /auth/refresh-token [post]
func RefreshTokenHandler(authService *services.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.RefreshTokenReq) {
		token, err := authService.RefreshToken(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "refresh token failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ToTokenResponseRet(token))
	})
}
