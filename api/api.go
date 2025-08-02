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
	"github.com/superwhys/rent-manager-server/api/middleware"
	"github.com/superwhys/rent-manager-server/api/permission"
	"github.com/superwhys/rent-manager-server/api/user"

	permissionRouter "github.com/superwhys/rent-manager-server/api/permission/routers"
	userRouter "github.com/superwhys/rent-manager-server/api/user/routers"
)

type handlerApp interface {
	permissionRouter.App
	userRouter.App
}

func SetupRouter(handlerApp handlerApp) *gin.Engine {
	middlewareManager := middleware.NewManager()

	engine := ginutils.NewServerHandler(
		ginutils.WithLoggingRequest(true),
		permission.Router(handlerApp, middlewareManager),
		user.Router(handlerApp, middlewareManager),
	)

	return engine
}
