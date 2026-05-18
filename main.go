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
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/superwhys/rent-manager-server/api"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
)

var (
	mysqlConfFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
)

func main() {
	flags.Parse()

	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfFlag(mysqlConf))

	client, err := repo.OpenMysqlClient(mysqlConf)
	logging.PanicError(err)
	defer client.Close()

	logging.PanicError(repo.MigrateSchema(client))

	rbacRepo := repo.NewRBACRepository(client)
	rbacService := services.NewRBACService(rbacRepo)
	router := api.SetupRouter(rbacService)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
