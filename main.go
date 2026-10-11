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

	"github.com/miebyte/authkit"
	authmysql "github.com/miebyte/authkit/mysql"
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"

	"github.com/superwhys/rent-manager/api"
	"github.com/superwhys/rent-manager/config"
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
	"github.com/superwhys/rent-manager/internal/infra/emailcode"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql"
	"github.com/superwhys/rent-manager/internal/infra/ratelimit"
	"github.com/superwhys/rent-manager/web"
)

var (
	authConfigFlag  = flags.Struct("auth", (*config.AuthConfig)(nil), "authkit config")
	adminConfigFlag = flags.Struct("admin", (*config.AdminConfig)(nil), "admin config")
	mysqlConfFlag   = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
	redisConfFlag   = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
)

func main() {
	flags.Parse()

	authConfig := new(config.AuthConfig)
	logging.PanicError(authConfigFlag(authConfig))
	logging.PanicError(authConfig.Validate())

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

	logging.PanicError(mysql.EnsureBuiltinPermissionSchema(client))
	logging.PanicError(client.AutoMigrate(authmysql.Models()...))
	logging.PanicError(client.AutoMigrate(models.AllModels()...))

	httpLimiter, err := ratelimit.NewRedisLimiter(10, time.Second*10, redisClient, "rent_manager")
	logging.PanicError(err)

	repos := mysql.NewRepositoryFactory(client).Repositories()

	authStore, err := authmysql.NewStore(client)
	logging.PanicError(err)
	var codeSender authkit.CodeSender
	if authConfig.SMTP.Configured() {
		codeSender, err = emailcode.NewSender(authConfig.SMTP)
		logging.PanicError(err)
	}
	authService, err := appauth.NewService(authStore, mysql.NewAccountProfileRepository(client), codeSender, authkit.Config{
		SessionTTL: authConfig.SessionTTL,
	})
	logging.PanicError(err)
	rbacService := apprbac.NewService(repos.RBAC, authStore.Accounts(), repos.Unit, adminConfig)
	adminServices := &appadmin.Services{
		AccountSetting: appaccountsetting.NewService(repos.AccountSetting),
		Home: apphome.NewService(
			repos.Unit,
			repos.Room,
			repos.RentBill,
			repos.MeterReading,
			repos.RBAC,
		),
		RBAC:      rbacService,
		Unit:      appunit.NewService(repos.Unit, repos.Room, repos.RentBill, repos.RBAC),
		Room:      approom.NewService(repos.Room, repos.Unit, repos.RBAC),
		RoomType:  approomtype.NewService(repos.RoomType, repos.Unit, repos.RBAC),
		MeterItem: appmeteritem.NewService(repos.MeterItem, repos.Unit, repos.RBAC),
		MeterReading: appmeterreading.NewService(
			repos.MeterReading,
			repos.Room,
			repos.Unit,
			repos.RBAC,
		),
		RentBill: apprentbill.NewService(
			repos.RentBill,
			repos.Room,
			repos.Unit,
			repos.MeterReading,
			repos.RBAC,
		),
	}
	router := api.SetupRouter(httpLimiter, authService, adminServices, adminConfig)
	webHandler, err := web.NewHandler()
	logging.PanicError(err)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", router),
		cores.WithHttpHandler("/swagger", api.SwaggerRouter(false)),
		cores.WithHttpHandler("/", webHandler),
	)

	logging.PanicError(cores.Start(srv, "0.0.0.0:8080"))
}
