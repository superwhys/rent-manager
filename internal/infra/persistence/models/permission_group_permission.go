package models

type PermissionGroupPermission struct {
	PermissionGroupID int              `gorm:"column:permission_group_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	PermissionCode    string           `gorm:"column:permission_code;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;primaryKey"`
	PermissionGroup   *PermissionGroup `gorm:"foreignKey:PermissionGroupID;constraint:permission_group_permissions_permission_group_id,OnDelete:CASCADE"`
}

func (PermissionGroupPermission) TableName() string {
	return "permission_group_permissions"
}
