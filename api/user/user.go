// File:		user.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package user

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/api/middleware"
	"github.com/superwhys/rent-manager-server/api/user/routers"
)

func Router(app routers.App, middlewareManager middleware.Manager) ginutils.EngineOption {
	userRouter := ginutils.NewRouter()
	userRouter.NewHandlerGroup().
		WithHandlers(
			ginutils.HandlerPath(http.MethodPost, "/login", ginutils.MountHandlerWithArg[routers.LoginHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/register", ginutils.MountHandlerWithArg[routers.RegisterHandler](app)),
		)

	// TODO: need auth middleware
	userRouter.NewHandlerGroup().
		WithMiddlewares(middlewareManager.Auth()).
		WithHandlers(
			ginutils.HandlerPath(http.MethodPost, "/logout", ginutils.MountHandlerWithArg[routers.LogoutHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/change-password", ginutils.MountHandlerWithArg[routers.ChangePasswordHandler](app)),
			ginutils.HandlerPath(http.MethodGet, "/info", ginutils.MountHandlerWithArg[routers.GetUserInfoHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/update", ginutils.MountHandlerWithArg[routers.UpdateUserInfoHandler](app)),
			ginutils.HandlerPath(http.MethodPost, "/delete", ginutils.MountHandlerWithArg[routers.DeleteUserHandler](app)),
		)

	return ginutils.WithGroupRouters(
		"/user",
		ginutils.WithRootRouters(userRouter),
	)
}
