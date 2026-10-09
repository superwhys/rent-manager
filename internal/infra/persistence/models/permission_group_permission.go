package models

type PermissionGroupPermission struct {
	PermissionGroupID int              `gorm:"column:permission_group_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	PermissionID      int              `gorm:"column:permission_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	PermissionGroup   *PermissionGroup `gorm:"foreignKey:PermissionGroupID;constraint:permission_group_permissions_permission_group_id,OnDelete:CASCADE"`
	Permission        *Permission      `gorm:"foreignKey:PermissionID;constraint:permission_group_permissions_permission_id,OnDelete:CASCADE"`
}

func (PermissionGroupPermission) TableName() string {
	return "permission_group_permissions"
}
