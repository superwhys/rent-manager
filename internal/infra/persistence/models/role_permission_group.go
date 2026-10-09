package models

type RolePermissionGroup struct {
	RoleID            int              `gorm:"column:role_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	PermissionGroupID int              `gorm:"column:permission_group_id;type:bigint;not null;primaryKey;autoIncrement:false"`
	Role              *Role            `gorm:"foreignKey:RoleID;constraint:role_permission_groups_role_id,OnDelete:CASCADE"`
	PermissionGroup   *PermissionGroup `gorm:"foreignKey:PermissionGroupID;constraint:role_permission_groups_permission_group_id,OnDelete:CASCADE"`
}

func (RolePermissionGroup) TableName() string {
	return "role_permission_groups"
}
