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
	"github.com/superwhys/rent-manager-server/api/permission"
)

type handlerApp interface {
	permission.App
}

func SetupRouter(handlerApp handlerApp) *gin.Engine {
	engine := ginutils.NewServerHandler(
		ginutils.WithLoggingRequest(true),
		permission.Router(handlerApp),
	)

	return engine
}
