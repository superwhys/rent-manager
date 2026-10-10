package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/logging"

	"github.com/superwhys/rent-manager/internal/xerr"
)

// HandleAuthError 在 HTTP 边界把 authkit 的稳定错误转换为中文响应。
func HandleAuthError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	status, responseErr := http.StatusInternalServerError, xerr.ErrSysInternal
	switch {
	case errors.Is(err, authkit.ErrInvalidInput):
		status, responseErr = http.StatusBadRequest, xerr.ErrBadRequest.WithMessage("登录参数无效")
	case errors.Is(err, authkit.ErrInvalidEmail):
		status, responseErr = http.StatusBadRequest, xerr.ErrBadRequest.WithMessage("邮箱格式不正确")
	case errors.Is(err, authkit.ErrInvalidCredentials):
		status, responseErr = http.StatusUnauthorized, xerr.ErrUnauthorized.WithMessage("邮箱或密码错误")
	case errors.Is(err, authkit.ErrUnauthorized):
		status, responseErr = http.StatusUnauthorized, xerr.ErrInvalidToken
	case errors.Is(err, authkit.ErrBlacklisted):
		status, responseErr = http.StatusForbidden, xerr.ErrUnauthorized.WithMessage("该账号暂不允许登录")
	case errors.Is(err, authkit.ErrRegistrationDenied):
		status, responseErr = http.StatusForbidden, xerr.ErrUnauthorized.WithMessage("该邮箱尚未注册，请先注册")
	case errors.Is(err, authkit.ErrConflict):
		status, responseErr = http.StatusConflict, xerr.ErrBadRequest.WithMessage("该邮箱已注册，请直接登录")
	case errors.Is(err, authkit.ErrTooManyRequests):
		status, responseErr = http.StatusTooManyRequests, xerr.ErrTooManyRequests
	case errors.Is(err, authkit.ErrResendTooSoon):
		status, responseErr = http.StatusTooManyRequests, xerr.ErrTooManyRequests.WithMessage("验证码发送过于频繁，请稍后重试")
	case errors.Is(err, authkit.ErrChallengeMismatch):
		status, responseErr = http.StatusBadRequest, xerr.ErrBadRequest.WithMessage("验证码错误，请重新输入")
	case errors.Is(err, authkit.ErrChallengeInvalid):
		status, responseErr = http.StatusBadRequest, xerr.ErrBadRequest.WithMessage("验证码已失效，请重新获取")
	case errors.Is(err, authkit.ErrChallengeUpdated):
		status, responseErr = http.StatusConflict, xerr.ErrBadRequest.WithMessage("验证码已更新，请使用最新验证码")
	case errors.Is(err, authkit.ErrEmailUnavailable):
		status, responseErr = http.StatusServiceUnavailable, xerr.ErrSysInternal.WithMessage("邮箱验证码服务尚未配置")
	case errors.Is(err, authkit.ErrMailFailed):
		status, responseErr = http.StatusBadGateway, xerr.ErrSysInternal.WithMessage("验证码发送失败，请稍后重试")
	default:
		logging.Errorw(ctx.Request.Context(), "authkit request failed", logging.Err(err))
	}
	ctx.JSON(status, ErrorResponseWithCode(responseErr))
	return true
}
