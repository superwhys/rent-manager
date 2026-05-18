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
func SetupRouter(
	httpLimiter *limiter.Limiter,
	jwtConfig *config.JwtConfig,
	authService *services.AuthService,
	adminServices *services.AdminServices,
) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(middlewares.RateLimitMiddleware(httpLimiter, nil)),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/auth"),
			routers.AuthRouter(authService),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/admin"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.AdminRouter(adminServices.RBAC),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/units"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.UnitRouter(adminServices.Unit, adminServices.Room, adminServices.RBAC),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/rooms"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.RoomRouter(adminServices.Room),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/room-types"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.RoomTypeRouter(adminServices.RoomType),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/meter-items"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.MeterItemRouter(adminServices.MeterItem),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/roles"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(jwtConfig)),
			routers.RoleRouter(adminServices.RBAC),
		),
	)

	return engine
}
