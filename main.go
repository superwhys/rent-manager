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

	"github.com/superwhys/rent-manager/api"
	"github.com/superwhys/rent-manager/config"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql"
	"github.com/superwhys/rent-manager/internal/infra/ratelimit"

	appaccount "github.com/superwhys/rent-manager/internal/app/account"
	appaccountsetting "github.com/superwhys/rent-manager/internal/app/accountsetting"
	appadmin "github.com/superwhys/rent-manager/internal/app/admin"
	appauth "github.com/superwhys/rent-manager/internal/app/auth"
	apphome "github.com/superwhys/rent-manager/internal/app/home"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	appmeterreading "github.com/superwhys/rent-manager/internal/app/meterreading"
	apprbac "github.com/superwhys/rent-manager/internal/app/rbac"
	apprentbill "github.com/superwhys/rent-manager/internal/app/rentbill"
	approom "github.com/superwhys/rent-manager/internal/app/room"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	appunit "github.com/superwhys/rent-manager/internal/app/unit"
	infraauthjwt "github.com/superwhys/rent-manager/internal/infra/auth/jwt"
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
	sqlDB, err := client.DB()
	logging.PanicError(err)
	defer sqlDB.Close()

	logging.PanicError(client.AutoMigrate(models.AllModels()...))

	httpLimiter, err := ratelimit.NewRedisLimiter(10, time.Second*10, redisClient, "rent_manager")
	logging.PanicError(err)

	repos := mysql.NewRepositoryFactory(client).Repositories()

	tokenManager := infraauthjwt.NewManager(jwtConfig)
	authService := appauth.NewService(wechatConfig, repos.Account, tokenManager)
	rbacService := apprbac.NewService(repos.RBAC, adminConfig)
	adminServices := &appadmin.Services{
		Account:        appaccount.NewService(repos.Account),
		AccountSetting: appaccountsetting.NewService(repos.AccountSetting),
		Home:           apphome.NewService(repos.Unit, repos.Room, repos.RentBill, repos.MeterReading, repos.RBAC),
		RBAC:           rbacService,
		Unit:           appunit.NewService(repos.Unit, repos.Room, repos.RBAC),
		Room:           approom.NewService(repos.Room, repos.Unit, repos.RBAC),
		RoomType:       approomtype.NewService(repos.RoomType),
		MeterItem:      appmeteritem.NewService(repos.MeterItem),
		MeterReading:   appmeterreading.NewService(repos.MeterReading, repos.Room, repos.Unit, repos.RBAC),
		RentBill:       apprentbill.NewService(repos.RentBill, repos.Room, repos.Unit, repos.MeterReading, repos.RBAC),
	}
	router := api.SetupRouter(httpLimiter, tokenManager, authService, adminServices)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
		cores.WithHttpHandler("/swagger", api.SwaggerRouter(false)),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
