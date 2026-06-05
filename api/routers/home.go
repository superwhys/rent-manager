package routers

import (
	"net/http"

	apiresponse "git.superwhys.top/rent-manager/rent-manager-server/api/response"
	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	apphome "git.superwhys.top/rent-manager/rent-manager-server/internal/app/home"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

// HomeRouter 注册首页聚合数据路由。
func HomeRouter(homeService *apphome.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "", GetHomeHandler(homeService)),
	)
}

// GetHomeHandler godoc
// @Summary 获取首页聚合数据
// @Tags home
// @Produce json
// @Success 200 {object} ginutils.Ret[home.Response]
// @Router /home [get]
func GetHomeHandler(homeService *apphome.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		summary, err := homeService.GetHomeSummary(c.Request.Context())
		if apiresponse.HandleRouterError(c, err, "get home failed", xerr.ErrSysInternal) {
			return
		}

		c.JSON(http.StatusOK, apiresponse.ResponseWithData(apphome.ToSummaryResponse(summary)))
	})
}
