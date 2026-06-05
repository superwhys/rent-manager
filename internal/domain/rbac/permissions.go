package rbac

const (
	PermissionUnitView   = "unit.view"
	PermissionUnitUpdate = "unit.update"

	PermissionRoomView   = "room.view"
	PermissionRoomManage = "room.manage"

	PermissionRoomTypeView   = "room_type.view"
	PermissionRoomTypeManage = "room_type.manage"

	PermissionMeterItemView   = "meter_item.view"
	PermissionMeterItemManage = "meter_item.manage"

	PermissionMeterReadingView   = "meter_reading.view"
	PermissionMeterReadingManage = "meter_reading.manage"

	PermissionRentBillView   = "rent_bill.view"
	PermissionRentBillManage = "rent_bill.manage"

	SystemRoleSuperAdmin                = "SuperAdmin"
	SystemPermissionAllowAllCode        = "SYSTEM_ALLOW_ALL"
	SystemPermissionAllowAllDescription = "Allow all APIs"
)
