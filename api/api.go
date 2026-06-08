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
	"github.com/ulule/limiter/v3"

	"git.superwhys.top/rent-manager/rent-manager-server/api/middlewares"
	"git.superwhys.top/rent-manager/rent-manager-server/api/routers"
	appadmin "git.superwhys.top/rent-manager/rent-manager-server/internal/app/admin"
	appauth "git.superwhys.top/rent-manager/rent-manager-server/internal/app/auth"
)

// SetupRouter godoc
// @title Rent Manager Server API
// @version 1.0
// @description Rent Manager Server
// @BasePath /api
func SetupRouter(
	httpLimiter *limiter.Limiter,
	tokenManager appauth.TokenManager,
	authService *appauth.Service,
	adminServices *appadmin.Services,
) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(middlewares.RateLimitMiddleware(httpLimiter, nil)),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/auth"),
			routers.AuthRouter(authService),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/accout"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.AccountRouter(adminServices.Account),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/home"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.HomeRouter(adminServices.Home),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/admin"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.AdminRouter(adminServices.RBAC),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/units"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.UnitRouter(adminServices.Unit, adminServices.Room, adminServices.RBAC),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/rooms"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.RoomRouter(adminServices.Room),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/room-types"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.RoomTypeRouter(adminServices.RoomType),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/meter-items"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.MeterItemRouter(adminServices.MeterItem),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/roles"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(tokenManager)),
			routers.RoleRouter(adminServices.RBAC),
		),
	)

	return engine
}
