// File:		main.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/rent-manager-server/api"
	"github.com/superwhys/rent-manager-server/internal/app/services"
)

func main() {
	appService := services.NewAppService()
	router := api.SetupRouter(appService)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
