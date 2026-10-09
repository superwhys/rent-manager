package models

import "time"

type Role struct {
	ID               int                `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt        time.Time          `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt        time.Time          `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt        *time.Time         `gorm:"column:deleted_at;type:timestamp"`
	Name             string             `gorm:"column:name;type:varchar(255);not null;uniqueIndex:role_account_id_name,priority:2"`
	IsDefaultOwner   bool               `gorm:"column:is_default_owner;type:boolean;not null;default:false"`
	IsActive         *bool              `gorm:"column:is_active;type:boolean;not null;default:true"`
	AccountID        string             `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:role_account_id_name,priority:1"`
	PermissionGroups []*PermissionGroup `gorm:"many2many:role_permission_groups;joinForeignKey:RoleID;joinReferences:PermissionGroupID;-:migration"`
}

func (Role) TableName() string {
	return "roles"
}
