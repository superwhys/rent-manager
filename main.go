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

	"git.superwhys.top/rent-manager/rent-manager-server/api"
	"git.superwhys.top/rent-manager/rent-manager-server/config"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/mysql"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/ratelimit"

	appaccount "git.superwhys.top/rent-manager/rent-manager-server/internal/app/account"
	appadmin "git.superwhys.top/rent-manager/rent-manager-server/internal/app/admin"
	appauth "git.superwhys.top/rent-manager/rent-manager-server/internal/app/auth"
	apphome "git.superwhys.top/rent-manager/rent-manager-server/internal/app/home"
	appmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/app/meteritem"
	apprbac "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rbac"
	approom "git.superwhys.top/rent-manager/rent-manager-server/internal/app/room"
	approomtype "git.superwhys.top/rent-manager/rent-manager-server/internal/app/roomtype"
	appunit "git.superwhys.top/rent-manager/rent-manager-server/internal/app/unit"
	infraauthjwt "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/auth/jwt"
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

	client, err := mysql.OpenMysqlClient(mysqlConf)
	logging.PanicError(err)
	defer client.Close()

	logging.PanicError(mysql.MigrateSchema(client))

	httpLimiter, err := ratelimit.NewRedisLimiter(10, time.Second*10, redisClient, "rent_manager")
	logging.PanicError(err)

	accountRepo := mysql.NewAccountRepository(client)
	rbacRepo := mysql.NewRBACRepository(client)
	unitRepo := mysql.NewUnitRepository(client)
	roomRepo := mysql.NewRoomRepository(client)
	roomTypeRepo := mysql.NewRoomTypeRepository(client)
	meterItemRepo := mysql.NewMeterItemRepository(client)
	rentBillRepo := mysql.NewRentBillRepository(client)

	tokenManager := infraauthjwt.NewManager(jwtConfig)
	authService := appauth.NewService(wechatConfig, accountRepo, tokenManager)
	rbacService := apprbac.NewService(rbacRepo, adminConfig)
	adminServices := &appadmin.Services{
		Account:   appaccount.NewService(accountRepo),
		Home:      apphome.NewService(unitRepo, roomRepo, rentBillRepo),
		RBAC:      rbacService,
		Unit:      appunit.NewService(unitRepo, roomRepo, rbacRepo),
		Room:      approom.NewService(roomRepo, unitRepo, rbacRepo),
		RoomType:  approomtype.NewService(roomTypeRepo),
		MeterItem: appmeteritem.NewService(meterItemRepo),
	}
	router := api.SetupRouter(httpLimiter, tokenManager, authService, adminServices)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
