package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

func AppRouter(appService *services.AppService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "/health", HealthHandler(appService)),
	)
}

// HealthHandler godoc
// @Summary 服务健康检查
// @Tags app
// @Produce json
// @Success 200 {object} ginutils.Ret[string]
// @Router /health [get]
func HealthHandler(appService *services.AppService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		c.JSON(http.StatusOK, response.ResponseWithData(appService.Health(c.Request.Context())))
	})
}
