package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"

	apiresponse "github.com/superwhys/rent-manager/api/response"
	appcore "github.com/superwhys/rent-manager/internal/app"
	appaccountsetting "github.com/superwhys/rent-manager/internal/app/accountsetting"
	"github.com/superwhys/rent-manager/internal/xerr"
)

// AccountSettingRouter 注册当前账号配置路由。
func AccountSettingRouter(settingService *appaccountsetting.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithHandler(http.MethodGet, "", GetSettingsHandler(settingService)),
		ginutils.WithHandler(http.MethodPut, "", UpdateSettingsHandler(settingService)),
	)
}

// GetSettingsHandler godoc
// @Summary 获取当前账号配置
// @Tags settings
// @Produce json
// @Success 200 {object} ginutils.Ret[accountsetting.SettingsResponse]
// @Security BearerAuth
// @Router /settings [get]
func GetSettingsHandler(settingService *appaccountsetting.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *appcore.EmptyRequest) {
		settings, err := settingService.GetSettings(c.Request.Context())
		if apiresponse.HandleRouterError(
			c,
			err,
			"get account settings failed",
			xerr.ErrSysInternal,
		) {
			return
		}
		c.JSON(
			http.StatusOK,
			apiresponse.ResponseWithData(appaccountsetting.ToSettingsResponse(settings)),
		)
	})
}

// UpdateSettingsHandler godoc
// @Summary 更新当前账号配置
// @Tags settings
// @Accept json
// @Produce json
// @Param request body accountsetting.UpdateSettingsCommand true "账号配置"
// @Success 200 {object} ginutils.Ret[accountsetting.SettingsResponse]
// @Security BearerAuth
// @Router /settings [put]
func UpdateSettingsHandler(settingService *appaccountsetting.Service) gin.HandlerFunc {
	return ginutils.RequestHandler(
		func(c *gin.Context, req *appaccountsetting.UpdateSettingsCommand) {
			settings, err := settingService.UpdateSettings(c.Request.Context(), req)
			if apiresponse.HandleRouterError(
				c,
				err,
				"update account settings failed",
				xerr.ErrSysInternal,
			) {
				return
			}
			c.JSON(
				http.StatusOK,
				apiresponse.ResponseWithData(appaccountsetting.ToSettingsResponse(settings)),
			)
		},
	)
}
