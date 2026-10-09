package admin

import (
	appaccountsetting "github.com/superwhys/rent-manager/internal/app/accountsetting"
	apphome "github.com/superwhys/rent-manager/internal/app/home"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	appmeterreading "github.com/superwhys/rent-manager/internal/app/meterreading"
	apprbac "github.com/superwhys/rent-manager/internal/app/rbac"
	apprentbill "github.com/superwhys/rent-manager/internal/app/rentbill"
	approom "github.com/superwhys/rent-manager/internal/app/room"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	appunit "github.com/superwhys/rent-manager/internal/app/unit"
)

// Services 聚合后台路由需要的应用服务。
type Services struct {
	AccountSetting *appaccountsetting.Service
	Home           *apphome.Service
	RBAC           *apprbac.Service
	Unit           *appunit.Service
	Room           *approom.Service
	RoomType       *approomtype.Service
	MeterItem      *appmeteritem.Service
	MeterReading   *appmeterreading.Service
	RentBill       *apprentbill.Service
}
