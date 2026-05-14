// File:		api.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/api/middlewares"
	"github.com/superwhys/rent-manager-server/api/routers"
	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/ulule/limiter/v3"
)

// SetupRouter godoc
// @title Rent Manager Server API
// @version 1.0
// @description Rent Manager Server
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupRouter(
	jwtConfig *config.JwtConfig,
	httpLimiter *limiter.Limiter,
	authService *services.AuthService,
	roomService *services.RoomService,
	buildingService *services.BuildingService,
	meterItemService *services.MeterItemService,
	organizationService *services.OrganizationService,
	rbacService *services.RBACService,
) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(middlewares.RateLimitMiddleware(httpLimiter, nil)),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/admin"),
			ginutils.WithMiddleware(
				middlewares.TokenVerifyMiddleware(jwtConfig),
				middlewares.SystemAdminMiddleware(rbacService),
			),
			routers.AdminGroupRouter(organizationService, rbacService),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/auth"),
			routers.AuthGroupRouter(authService),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(
				middlewares.TokenVerifyMiddleware(jwtConfig),
				middlewares.PermissionMiddleware(rbacService),
			),
			routers.BuildingGroupRouter(buildingService),
			routers.RoomGroupRouter(roomService),
			routers.MeterItemGroupRouter(meterItemService),
			routers.RoleGroupRouter(rbacService),
		),
	)

	return engine
}
