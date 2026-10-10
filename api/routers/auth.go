package routers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/middlewares"
	apiresponse "github.com/superwhys/rent-manager/api/response"
	"github.com/superwhys/rent-manager/config"
	appauth "github.com/superwhys/rent-manager/internal/app/auth"
	"github.com/superwhys/rent-manager/internal/xerr"
)

type AuthService interface {
	middlewares.SessionAuthenticator
	SendCode(context.Context, authkit.SendCodeInput) error
	LoginEmail(context.Context, authkit.EmailLoginInput) (*authkit.LoginResult, error)
	LoginPassword(context.Context, authkit.PasswordLoginInput) (*authkit.LoginResult, error)
	Register(context.Context, appauth.RegisterRequest, string) (*authkit.LoginResult, error)
	UpdateAccount(context.Context, appauth.UpdateAccountRequest) (*authkit.Account, error)
	Logout(context.Context, string) error
}

type LoginResponse struct {
	Account      authkit.Account `json:"account"`
	AccessToken  string          `json:"access_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
	Created      bool            `json:"created"`
	IsSuperAdmin bool            `json:"is_super_admin"`
}

type SessionResponse struct {
	Account      authkit.Account `json:"account"`
	ExpiresAt    time.Time       `json:"expires_at"`
	IsSuperAdmin bool            `json:"is_super_admin"`
}

// AuthRouter 将认证接口直接交给 authkit。
func AuthRouter(authService AuthService, adminConfig *config.AdminConfig) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithMiddleware(func(c *gin.Context) {
			c.Header("Cache-Control", "no-store")
			c.Next()
		}),
		ginutils.WithHandler(http.MethodPost, "/email/code", SendEmailCodeHandler(authService)),
		ginutils.WithHandler(http.MethodPost, "/register", RegisterHandler(authService, adminConfig)),
		ginutils.WithHandler(http.MethodPost, "/email-login", EmailLoginHandler(authService, adminConfig)),
		ginutils.WithHandler(http.MethodPost, "/password-login", PasswordLoginHandler(authService, adminConfig)),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			ginutils.WithHandler(http.MethodPatch, "/account", UpdateAccountHandler(authService)),
			ginutils.WithHandler(http.MethodGet, "/session", SessionHandler(adminConfig)),
			ginutils.WithHandler(http.MethodDelete, "/session", LogoutHandler(authService)),
		),
	)
}

// UpdateAccountHandler godoc
// @Summary 修改当前账号名称
// @Tags auth
// @Accept json
// @Produce json
// @Param request body appauth.UpdateAccountRequest true "账号名称（去除首尾空白后为 1–64 个字符，不允许控制字符）"
// @Success 200 {object} ginutils.Ret[authkit.Account]
// @Security BearerAuth
// @Router /auth/account [patch]
func UpdateAccountHandler(authService AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.UpdateAccountRequest) {
		account, err := authService.UpdateAccount(c.Request.Context(), *req)
		if errors.Is(err, authkit.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, apiresponse.ErrorResponseWithCode(
				xerr.ErrBadRequest.WithMessage("账号名称须为 1–64 个字符，且不能包含控制字符"),
			))
			return
		}
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(account))
	})
}

// SendEmailCodeHandler godoc
// @Summary 发送邮箱验证码
// @Tags auth
// @Accept json
// @Produce json
// @Param request body appauth.SendEmailCodeRequest true "接收验证码的邮箱"
// @Success 200 {object} ginutils.Ret[appauth.SendEmailCodeResponse]
// @Router /auth/email/code [post]
func SendEmailCodeHandler(authService AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.SendEmailCodeRequest) {
		err := authService.SendCode(c.Request.Context(), authkit.SendCodeInput{
			Email: req.Email, IP: c.ClientIP(),
		})

		if apiresponse.HandleAuthError(c, err) {
			return
		}

		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appauth.SendEmailCodeResponse{
			ResendAfterSeconds: int(authkit.ResendInterval / time.Second),
			ExpiresInSeconds:   int(authkit.CodeTTL / time.Second),
		}))
	})
}

// RegisterHandler godoc
// @Summary 验证邮箱并注册密码账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body appauth.RegisterRequest true "邮箱注册参数"
// @Success 200 {object} ginutils.Ret[LoginResponse]
// @Router /auth/register [post]
func RegisterHandler(authService AuthService, adminConfig *config.AdminConfig) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.RegisterRequest) {
		login, err := authService.Register(c.Request.Context(), *req, c.ClientIP())
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		writeLoginResponse(c, login, adminConfig)
	})
}

// EmailLoginHandler godoc
// @Summary 使用邮箱验证码登录已有账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body appauth.EmailLoginRequest true "验证码登录参数"
// @Success 200 {object} ginutils.Ret[LoginResponse]
// @Router /auth/email-login [post]
func EmailLoginHandler(authService AuthService, adminConfig *config.AdminConfig) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.EmailLoginRequest) {
		login, err := authService.LoginEmail(
			c.Request.Context(),
			authkit.EmailLoginInput{Email: req.Email, Code: req.Code},
		)
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		writeLoginResponse(c, login, adminConfig)
	})
}

// PasswordLoginHandler godoc
// @Summary 使用邮箱和密码登录已有账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body appauth.PasswordLoginRequest true "密码登录参数"
// @Success 200 {object} ginutils.Ret[LoginResponse]
// @Router /auth/password-login [post]
func PasswordLoginHandler(authService AuthService, adminConfig *config.AdminConfig) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appauth.PasswordLoginRequest) {
		email, err := authkit.NormalizeEmail(req.Email)
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		login, err := authService.LoginPassword(c.Request.Context(), authkit.PasswordLoginInput{
			Identifier: email, Password: req.Password, IP: c.ClientIP(),
		})
		if apiresponse.HandleAuthError(c, err) {
			return
		}
		writeLoginResponse(c, login, adminConfig)
	})
}

func writeLoginResponse(c *gin.Context, login *authkit.LoginResult, adminConfig *config.AdminConfig) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, apiresponse.ResponseWithData(LoginResponse{
		Account:      login.Account,
		AccessToken:  login.Token,
		ExpiresAt:    login.Expires,
		Created:      login.Created,
		IsSuperAdmin: adminConfig != nil && adminConfig.IsSuperAdmin(login.Account.ID),
	}))
}

// SessionHandler godoc
// @Summary 获取当前 authkit 账号与会话有效期
// @Tags auth
// @Produce json
// @Success 200 {object} ginutils.Ret[SessionResponse]
// @Security BearerAuth
// @Router /auth/session [get]
func SessionHandler(adminConfig *config.AdminConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := middlewares.AuthenticatedSessionFromContext(c)
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(SessionResponse{
			Account:      session.Account,
			ExpiresAt:    session.Expires,
			IsSuperAdmin: adminConfig != nil && adminConfig.IsSuperAdmin(session.Account.ID),
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
		if apiresponse.HandleAuthError(
			c,
			authService.Logout(c.Request.Context(), middlewares.BearerToken(c)),
		) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseSuccess())
	}
}
