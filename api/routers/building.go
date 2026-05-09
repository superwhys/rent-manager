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

func BuildingGroupRouter(buildingService *services.BuildingService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/buildings"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/", BuildingCreateHandler(buildingService)),
			ginutils.WithHandler(http.MethodGet, "/", BuildingListHandler(buildingService)),
			ginutils.WithHandler(http.MethodGet, "/:building_id", BuildingGetHandler(buildingService)),
			ginutils.WithHandler(http.MethodPut, "/:building_id", BuildingUpdateHandler(buildingService)),
			ginutils.WithHandler(http.MethodDelete, "/:building_id", BuildingDeleteHandler(buildingService)),
			ginutils.WithHandler(http.MethodPost, "/:building_id/rooms", BuildingCreateRoomHandler(buildingService)),
		),
	)
}

// BuildingCreateHandler 处理创建楼栋请求。
// @Summary 创建楼栋
// @Description 创建楼栋并设置默认抄表项
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requests.CreateBuildingCommand true "创建楼栋请求体"
// @Success 200 {object} ginutils.Ret[response.BuildingResponse]
// @Router /buildings [post]
func BuildingCreateHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateBuildingCommand) {
		building, err := buildingService.CreateBuilding(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create building failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.BuildingResponseRet(building))
	})
}

// BuildingListHandler 处理楼栋列表查询请求。
// @Summary 查询楼栋列表
// @Description 查询楼栋列表
// @Tags Building
// @Produce json
// @Security BearerAuth
// @Param request query requests.ListBuildingsQuery true "获取楼栋列表请求参数"
// @Success 200 {object} ginutils.Ret[[]response.BuildingResponse]
// @Router /buildings [get]
func BuildingListHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.ListBuildingsQuery) {
		buildings, err := buildingService.ListBuildings(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "list buildings failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.BuildingListResponseRet(buildings))
	})
}

// BuildingGetHandler 处理楼栋详情查询请求。
// @Summary 查询楼栋详情
// @Description 查询楼栋详情
// @Tags Building
// @Produce json
// @Security BearerAuth
// @Param building_id path int true "楼栋 ID"
// @Success 200 {object} ginutils.Ret[response.BuildingResponse]
// @Router /buildings/{building_id} [get]
func BuildingGetHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.BuildingIDQuery) {
		building, err := buildingService.GetBuilding(c.Request.Context(), req.BuildingID)
		if response.HandleRouterError(c, err, "get building failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.BuildingResponseRet(building))
	})
}

// BuildingUpdateHandler 处理更新楼栋请求。
// @Summary 更新楼栋
// @Description 更新楼栋并整体替换默认抄表项
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param building_id path int true "楼栋 ID"
// @Param request body requests.UpdateBuildingCommand true "更新楼栋请求体"
// @Success 200 {object} ginutils.Ret[response.BuildingResponse]
// @Router /buildings/{building_id} [put]
func BuildingUpdateHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.UpdateBuildingCommand) {
		building, err := buildingService.UpdateBuilding(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "update building failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.BuildingResponseRet(building))
	})
}

// BuildingDeleteHandler 处理删除楼栋请求。
// @Summary 删除楼栋
// @Description 删除没有房间的楼栋
// @Tags Building
// @Produce json
// @Security BearerAuth
// @Param building_id path int true "楼栋 ID"
// @Success 200 {object} ginutils.Ret[any]
// @Router /buildings/{building_id} [delete]
func BuildingDeleteHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.BuildingIDQuery) {
		err := buildingService.DeleteBuilding(c.Request.Context(), req.BuildingID)
		if response.HandleRouterError(c, err, "delete building failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// BuildingCreateRoomHandler 处理在楼栋下创建房间请求。
// @Summary 在楼栋下创建房间
// @Description 在指定楼栋下创建房间
// @Tags Building
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param building_id path int true "楼栋 ID"
// @Param request body requests.CreateRoomInBuildingCommand true "创建房间请求体"
// @Success 200 {object} ginutils.Ret[response.RoomResponse]
// @Router /buildings/{building_id}/rooms [post]
func BuildingCreateRoomHandler(buildingService *services.BuildingService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *requests.CreateRoomInBuildingCommand) {
		room, err := buildingService.CreateRoomInBuilding(c.Request.Context(), req)
		if response.HandleRouterError(c, err, "create room in building failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.RoomResponseRet(room))
	})
}
