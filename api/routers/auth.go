package routers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/middlewares"
	apiresponse "github.com/superwhys/rent-manager/api/response"
)

type AuthService interface {
	middlewares.SessionAuthenticator
	LoginWechat(context.Context, string) (*authkit.LoginResult, error)
	Logout(context.Context, string) error
}

type WechatLoginRequest struct {
	Code string `json:"code" form:"-" header:"-" validate:"required"`
}

type LoginResponse struct {
	Account     authkit.Account `json:"account"`
	AccessToken string          `json:"access_token"`
	ExpiresAt   time.Time       `json:"expires_at"`
	Created     bool            `json:"created"`
}

type SessionResponse struct {
	Account   authkit.Account `json:"account"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// AuthRouter 将认证接口直接交给 authkit。
func AuthRouter(authService AuthService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodPost, "/wechat-login", WechatLoginHandler(authService)),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			ginutils.WithHandler(http.MethodGet, "/session", SessionHandler()),
			ginutils.WithHandler(http.MethodDelete, "/session", LogoutHandler(authService)),
		),
	)
}

// WechatLoginHandler godoc
// @Summary 微信登录并创建 authkit 会话
// @Tags auth
// @Accept json
// @Produce json
// @Param request body WechatLoginRequest true "微信登录参数"
// @Success 200 {object} ginutils.Ret[LoginResponse]
// @Router /auth/wechat-login [post]
func WechatLoginHandler(authService AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *WechatLoginRequest) {
		login, err := authService.LoginWechat(c.Request.Context(), req.Code)
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(LoginResponse{
			Account: login.Account, AccessToken: login.Token, ExpiresAt: login.Expires, Created: login.Created,
		}))
	})
}

// SessionHandler godoc
// @Summary 获取当前 authkit 账号与会话有效期
// @Tags auth
// @Produce json
// @Success 200 {object} ginutils.Ret[SessionResponse]
// @Security BearerAuth
// @Router /auth/session [get]
func SessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := middlewares.AuthenticatedSessionFromContext(c)
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(SessionResponse{
			Account: session.Account, ExpiresAt: session.Expires,
		}))
	}
}

// LogoutHandler godoc
// @Summary 撤销当前 authkit 会话
// @Tags auth
// @Produce json
// @Success 200 {object} ginutils.Ret[any]
// @Security BearerAuth
// @Router /auth/session [delete]
func LogoutHandler(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiresponse.HandleAuthError(c, authService.Logout(c.Request.Context(), middlewares.BearerToken(c))) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	}
}
