// File:		permission.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package permission

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/api/middleware"
	"github.com/superwhys/rent-manager-server/api/permission/routers"
)

func Router(app routers.App, middlewareManager middleware.Manager) ginutils.EngineOption {
	permissionRouter := ginutils.NewRouter()
	permissionRouter.NewHandlerGroup().
		WithMiddlewares(middlewareManager.Auth()).
		WithHandlers(
			ginutils.HandlerPath(http.MethodGet, "/list", ginutils.MountHandlerWithArg[routers.ListPermissionHandler](app)),
		)

	return ginutils.WithGroupRouters(
		"/permission",
		ginutils.WithRootRouters(permissionRouter),
	)
}
