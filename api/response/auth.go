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
	case errors.Is(err, authkit.ErrWechatCode), errors.Is(err, authkit.ErrInvalidInput):
		status, responseErr = http.StatusBadRequest, xerr.ErrBadRequest.WithMessage("登录参数无效")
	case errors.Is(err, authkit.ErrUnauthorized), errors.Is(err, authkit.ErrInvalidCredentials):
		status, responseErr = http.StatusUnauthorized, xerr.ErrInvalidToken
	case errors.Is(err, authkit.ErrBlacklisted), errors.Is(err, authkit.ErrRegistrationDenied):
		status, responseErr = http.StatusForbidden, xerr.ErrUnauthorized.WithMessage("该账号暂不允许登录")
	case errors.Is(err, authkit.ErrTooManyRequests):
		status, responseErr = http.StatusTooManyRequests, xerr.ErrTooManyRequests
	case errors.Is(err, authkit.ErrWechatUnavailable):
		status, responseErr = http.StatusServiceUnavailable, xerr.ErrSysInternal.WithMessage("微信登录尚未配置")
	case errors.Is(err, authkit.ErrWechatLogin):
		status, responseErr = http.StatusBadGateway, xerr.ErrSysInternal.WithMessage("微信登录失败，请稍后重试")
	default:
		logging.Errorw(ctx.Request.Context(), "authkit request failed", logging.Err(err))
	}
	ctx.JSON(status, ErrorResponseWithCode(responseErr))
	return true
}
