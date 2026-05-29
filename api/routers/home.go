package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
	"github.com/superwhys/rent-manager-server/internal/models/response"
)

// HomeRouter 注册首页聚合数据路由。
func HomeRouter(homeService *services.HomeService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "", GetHomeHandler(homeService)),
	)
}

// GetHomeHandler godoc
// @Summary 获取首页聚合数据
// @Tags home
// @Produce json
// @Success 200 {object} ginutils.Ret[response.HomeResponse]
// @Router /home [get]
func GetHomeHandler(homeService *services.HomeService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.EmptyRequest) {
		summary, err := homeService.GetHomeSummary(c.Request.Context())
		if response.HandleRouterError(c, err, "get home failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.HomeResponseRet(response.HomeResponse{
			BuildingRoomSummary: response.HomeBuildingRoomSummaryResponse{
				BuildingCount:   summary.BuildingCount,
				RoomCount:       summary.RoomCount,
				RentedRoomCount: summary.RentedRoomCount,
				VacantRoomCount: summary.VacantRoomCount,
			},
			IncomeAmount:  "38,650",
			IncomeDecimal: ".00",
			MonthlyGrowth: "+¥1,280",
			CurrentMonth:  "2025-05",
			Stats: []response.HomeStatResponse{
				{
					Value: "5",
					Label: "楼栋",
					Tone:  "green",
				},
				{
					Value: "42",
					Label: "房间",
					Tone:  "blue",
				},
				{
					Value: "6",
					Label: "待收",
					Tone:  "orange",
				},
			},
			Todos: []response.HomeTodoResponse{
				{
					Title:       "幸福小区 · 3栋",
					Description: "还有 8 间房未抄表",
					Tag:         "抄表",
				},
				{
					Title:       "阳光花园 · 2栋",
					Description: "还有 6 笔租金待收",
					Tag:         "收租",
				},
				{
					Title:       "城市公寓 · 801",
					Description: "租约将在 12 天后到期",
					Tag:         "续租",
				},
			},
		}))
	})
}
