package rbac

const (
	PermissionUnitView   = "unit.view"
	PermissionUnitUpdate = "unit.update"

	PermissionRoomView   = "room.view"
	PermissionRoomManage = "room.manage"

	PermissionRoomTypeView = "room_type.view"
	// PermissionRoomTypeManage 兼容保留；单元授权不允许修改账号共享房型定义，定义仅由所有者管理。
	PermissionRoomTypeManage = "room_type.manage"

	PermissionMeterItemView = "meter_item.view"
	// PermissionMeterItemManage 兼容保留；单元授权不允许修改账号共享收费定义，定义仅由所有者管理。
	PermissionMeterItemManage = "meter_item.manage"

	PermissionMeterReadingView   = "meter_reading.view"
	PermissionMeterReadingManage = "meter_reading.manage"

	PermissionRentBillView   = "rent_bill.view"
	PermissionRentBillManage = "rent_bill.manage"

	SystemRoleSuperAdmin                = "SuperAdmin"
	SystemPermissionAllowAllCode        = "SYSTEM_ALLOW_ALL"
	SystemPermissionAllowAllDescription = "Allow all APIs"
)
