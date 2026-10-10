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

var builtinPermissions = [...]Permission{
	{Code: PermissionUnitView, Description: "查看楼栋详情"},
	{Code: PermissionUnitUpdate, Description: "修改楼栋资料"},
	{Code: PermissionRoomView, Description: "查看房间"},
	{Code: PermissionRoomManage, Description: "管理房间"},
	{Code: PermissionRoomTypeView, Description: "查看楼栋所有者的共享房型"},
	{Code: PermissionRentBillView, Description: "查看账单与财务汇总"},
	{Code: PermissionRentBillManage, Description: "生成、收款及取消账单"},
	{Code: PermissionMeterReadingView, Description: "查看月度抄表"},
	{Code: PermissionMeterReadingManage, Description: "保存月度读数"},
	{Code: PermissionMeterItemView, Description: "查看楼栋所有者的共享收费项目"},
}

// BuiltinPermissions 返回业务代码支持的固定权限目录副本。
func BuiltinPermissions() []*Permission {
	permissions := make([]*Permission, len(builtinPermissions))
	for i := range builtinPermissions {
		permission := builtinPermissions[i]
		permissions[i] = &permission
	}
	return permissions
}

// IsBuiltinPermission 按完整编码检查权限是否属于固定目录。
func IsBuiltinPermission(code string) bool {
	for _, permission := range builtinPermissions {
		if permission.Code == code {
			return true
		}
	}
	return false
}
