// File:		tenant.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package tenant

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/api/middleware"
	"github.com/superwhys/rent-manager-server/api/tenant/routers"
)

func Router(app routers.App, middlewareManager middleware.Manager) ginutils.EngineOption {
	tenantRouter := ginutils.NewRouter()

	tenantRouter.NewHandlerGroup().
		WithMiddlewares(middlewareManager.Auth()).
		WithHandlers(
			ginutils.HandlerPath(http.MethodGet, "/list", ginutils.MountHandlerWithArg[routers.ListTenantHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/create", ginutils.MountHandlerWithArg[routers.CreateTenantHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/update", ginutils.MountHandlerWithArg[routers.UpdateTenantHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/delete", ginutils.MountHandlerWithArg[routers.DeleteTenantHandler](app)),
		)

	return ginutils.WithGroupRouters(
		"/tenant",
		ginutils.WithRootRouters(tenantRouter),
	)
}
