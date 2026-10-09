// File:		api.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager/api/middlewares"
	"github.com/superwhys/rent-manager/api/routers"
	"github.com/ulule/limiter/v3"

	appadmin "github.com/superwhys/rent-manager/internal/app/admin"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/superwhys/rent-manager/api/docs"
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
	httpLimiter *limiter.Limiter,
	authService routers.AuthService,
	adminServices *appadmin.Services,
) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(middlewares.RateLimitMiddleware(httpLimiter, nil)),
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/auth"),
			routers.AuthRouter(authService),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/settings"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.AccountSettingRouter(adminServices.AccountSetting),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/home"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.HomeRouter(adminServices.Home),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/admin"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.AdminRouter(adminServices.RBAC),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/units"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.UnitRouter(adminServices.Unit, adminServices.Room, adminServices.RBAC, adminServices.MeterReading, adminServices.RentBill),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/rooms"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.RoomRouter(adminServices.Room, adminServices.MeterReading, adminServices.RentBill),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/rent-bills"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.RentBillRouter(adminServices.RentBill),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/room-types"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.RoomTypeRouter(adminServices.RoomType),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/meter-items"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.MeterItemRouter(adminServices.MeterItem),
		),

		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/roles"),
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(authService)),
			routers.RoleRouter(adminServices.RBAC),
		),
	)

	return engine
}

func SwaggerRouter(isProd bool) http.Handler {
	if isProd {
		return http.NotFoundHandler()
	}

	return httpSwagger.Handler()
}
