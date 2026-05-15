// File:		main.go
// Created by:	Hoven
// Created on:	2025-07-31
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"context"
	"time"

	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/rent-manager-server/api"
	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/infra/ratelimit"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
)

var (
	wechatConfigFlag = flags.Struct("wechat", (*config.WechatConfig)(nil), "wechat config")
	jwtConfigFlag    = flags.Struct("jwt", (*config.JwtConfig)(nil), "jwt config")
	redisConfigFlag  = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
	mysqlConfFlag    = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
)

func main() {
	flags.Parse()

	wechatConfig := new(config.WechatConfig)
	logging.PanicError(wechatConfigFlag(wechatConfig))

	jwtConfig := new(config.JwtConfig)
	logging.PanicError(jwtConfigFlag(jwtConfig))

	// 加载 redis 配置
	redisConf := new(redisutils.RedisConfig)
	logging.PanicError(redisConfigFlag(redisConf))

	redisClient, err := redisConf.DialGORedisClient()
	logging.PanicError(err)

	// 加载 mysql 配置
	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfFlag(mysqlConf))

	// 初始化 mysql 客户端
	client, err := repo.OpenMysqlClient(mysqlConf)
	logging.PanicError(err)
	defer client.Close()

	// 自动迁移数据库表结构
	// TODO: 需要在生产环境中禁用此功能
	logging.PanicError(repo.MigrateSchema(client))

	httpLimiter, err := ratelimit.NewRedisLimiter(10, time.Second*10, redisClient, "http_rate")
	logging.PanicError(err)

	roomRepo := repo.NewRoomRepository(client)
	unitRepo := repo.NewUnitRepository(client)
	meterItemRepo := repo.NewMeterItemRepository(client)
	accountRepo := repo.NewAccountRepository(client)
	rbacRepo := repo.NewRBACRepository(client)

	roomService := services.NewRoomService(roomRepo)
	unitService := services.NewUnitService(unitRepo, roomService)
	meterItemService := services.NewMeterItemService(meterItemRepo)
	authService := services.NewAuthService(jwtConfig, wechatConfig, accountRepo)
	rbacService := services.NewRBACService(rbacRepo, redisClient)
	logging.PanicError(rbacService.InitSystemRBAC(context.Background()))

	router := api.SetupRouter(
		jwtConfig,
		httpLimiter,
		authService,
		roomService,
		unitService,
		meterItemService,
		rbacService,
	)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
