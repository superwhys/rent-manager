// File:		main.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"time"

	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/rent-manager-server/api"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/infra/ratelimit"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
)

var (
	mysqlConfFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
	redisConfFlag = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
)

func main() {
	flags.Parse()

	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfFlag(mysqlConf))

	redisConf := new(redisutils.RedisConfig)
	logging.PanicError(redisConfFlag(redisConf))

	redisClient, err := redisConf.DialGORedisClient()
	logging.PanicError(err)

	client, err := repo.OpenMysqlClient(mysqlConf)
	logging.PanicError(err)
	defer client.Close()

	logging.PanicError(repo.MigrateSchema(client))

	httpLimiter, err := ratelimit.NewRedisLimiter(10, time.Second*10, redisClient, "rent_manager")
	logging.PanicError(err)

	rbacRepo := repo.NewRBACRepository(client)
	rbacService := services.NewRBACService(rbacRepo)
	router := api.SetupRouter(httpLimiter, rbacService)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
