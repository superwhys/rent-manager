package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "github.com/superwhys/rent-manager/api/response"
	appcore "github.com/superwhys/rent-manager/internal/app"
	appaccount "github.com/superwhys/rent-manager/internal/app/account"
	"github.com/superwhys/rent-manager/internal/xerr"
)

// AccountRouter 注册账号相关路由。
func AccountRouter(accountService *appaccount.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/self", AccountSelfHandler(accountService)),
	)
}

// AccountSelfHandler godoc
// @Summary 获取个人信息
// @Tags accout
// @Produce json
// @Success 200 {object} ginutils.Ret[account.AccountResponse]
// @Security BearerAuth
// @Router /accout/self [get]
func AccountSelfHandler(accountService *appaccount.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		accountInfo, err := accountService.Self(c.Request.Context())
		if apiresponse.HandleRouterError(c, err, "get account self failed", xerr.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, apiresponse.ResponseWithData(appaccount.ToAccountResponse(accountInfo)))
	})
}
