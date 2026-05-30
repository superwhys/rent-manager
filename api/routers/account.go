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

func AccountRouter(accountService *services.AccountService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/self", AccountSelfHandler(accountService)),
	)
}

// AccountSelfHandler godoc
// @Summary 获取个人信息
// @Tags accout
// @Produce json
// @Success 200 {object} ginutils.Ret[response.AccountResponse]
// @Router /accout/self [get]
func AccountSelfHandler(accountService *services.AccountService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		accountInfo, err := accountService.Self(c.Request.Context())
		if response.HandleRouterError(c, err, "get account self failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.ToAccountResponse(accountInfo)))
	})
}
