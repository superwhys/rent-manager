package admin

import (
	appaccount "git.superwhys.top/rent-manager/rent-manager-server/internal/app/account"
	apphome "git.superwhys.top/rent-manager/rent-manager-server/internal/app/home"
	appmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/app/meteritem"
	apprbac "git.superwhys.top/rent-manager/rent-manager-server/internal/app/rbac"
	approom "git.superwhys.top/rent-manager/rent-manager-server/internal/app/room"
	approomtype "git.superwhys.top/rent-manager/rent-manager-server/internal/app/roomtype"
	appunit "git.superwhys.top/rent-manager/rent-manager-server/internal/app/unit"
)

// Services 聚合后台路由需要的应用服务。
type Services struct {
	Account   *appaccount.Service
	Home      *apphome.Service
	RBAC      *apprbac.Service
	Unit      *appunit.Service
	Room      *approom.Service
	RoomType  *approomtype.Service
	MeterItem *appmeteritem.Service
}
