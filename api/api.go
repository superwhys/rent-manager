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
	"github.com/superwhys/rent-manager-server/api/routers"
	"github.com/superwhys/rent-manager-server/internal/app/services"
)

// SetupRouter godoc
// @title Rent Manager Server API
// @version 1.0
// @description Rent Manager Server
// @BasePath /api
func SetupRouter(
	rbacService *services.RBACService,
) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithGroupHandlers(
			ginutils.WithPrefix("/admin"),
			routers.AdminRouter(rbacService),
		),
	)

	return engine
}
