package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"

	"github.com/superwhys/rent-manager/internal/xerr"
)

// HandleRouterError 处理业务逻辑错误响应。
func HandleRouterError(ctx *gin.Context, err error, logMsg string, fallback xerr.Error) bool {
	if err == nil {
		return false
	}

	if ec, ok := xerr.AsErrcode(err); ok {
		ctx.JSON(http.StatusOK, ginutils.ErrorRet(ec))
		return true
	}

	logging.Errorw(ctx.Request.Context(), logMsg, logging.Err(err))
	ctx.JSON(http.StatusOK, ginutils.ErrorRet(fallback))
	return true
}
