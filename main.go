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
	"github.com/superwhys/rent-manager-server/config"
	"github.com/superwhys/rent-manager-server/internal/app/services"
	"github.com/superwhys/rent-manager-server/internal/infra/ratelimit"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
)

var (
	wechatConfigFlag = flags.Struct("wechat", (*config.WechatConfig)(nil), "wechat config")
	jwtConfigFlag    = flags.Struct("jwt", (*config.JwtConfig)(nil), "jwt config")
	adminConfigFlag  = flags.Struct("admin", (*config.AdminConfig)(nil), "admin config")
	mysqlConfFlag    = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
	redisConfFlag    = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
)

func main() {
	flags.Parse()

	wechatConfig := new(config.WechatConfig)
	logging.PanicError(wechatConfigFlag(wechatConfig))
	logging.PanicError(wechatConfig.Validate())

	jwtConfig := new(config.JwtConfig)
	logging.PanicError(jwtConfigFlag(jwtConfig))
	logging.PanicError(jwtConfig.Validate())

	adminConfig := new(config.AdminConfig)
	logging.PanicError(adminConfigFlag(adminConfig))

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

	accountRepo := repo.NewAccountRepository(client)
	rbacRepo := repo.NewRBACRepository(client)
	unitRepo := repo.NewUnitRepository(client)
	roomRepo := repo.NewRoomRepository(client)
	roomTypeRepo := repo.NewRoomTypeRepository(client)
	meterItemRepo := repo.NewMeterItemRepository(client)

	authService := services.NewAuthService(jwtConfig, wechatConfig, accountRepo)
	rbacService := services.NewRBACService(rbacRepo, adminConfig)
	adminServices := &services.AdminServices{
		Account:   services.NewAccountService(accountRepo),
		Home:      services.NewHomeService(unitRepo, roomRepo),
		RBAC:      rbacService,
		Unit:      services.NewUnitService(unitRepo, roomRepo, rbacService),
		Room:      services.NewRoomService(roomRepo, unitRepo, rbacService),
		RoomType:  services.NewRoomTypeService(roomTypeRepo),
		MeterItem: services.NewMeterItemService(meterItemRepo),
	}
	router := api.SetupRouter(httpLimiter, jwtConfig, authService, adminServices)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
