package models

// AllModels 包含业务表与关联表；多对多字段不参与迁移，由关联表 Model 保留原有外键。
func AllModels() []any {
	return []any{
		&Account{},
		&AccountSetting{},
		&MeterItem{},
		&MeterReading{},
		&Permission{},
		&PermissionGroup{},
		&RentBill{},
		&Role{},
		&Room{},
		&RoomType{},
		&Unit{},
		&UnitAccountRole{},
		&PermissionGroupPermission{},
		&RolePermissionGroup{},
		&RoomMeterItem{},
		&UnitMeterItem{},
	}
}
