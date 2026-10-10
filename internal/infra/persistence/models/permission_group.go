package models

import "time"

type PermissionGroup struct {
	ID                 int                          `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt          time.Time                    `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt          time.Time                    `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt          *time.Time                   `gorm:"column:deleted_at;type:timestamp"`
	Code               string                       `gorm:"column:code;type:varchar(255);not null;uniqueIndex:permissiongroup_code"`
	Name               string                       `gorm:"column:name;type:varchar(255);not null"`
	Description        *string                      `gorm:"column:description;type:varchar(255)"`
	SortOrder          int                          `gorm:"column:sort_order;type:bigint;not null;default:0"`
	IsActive           *bool                        `gorm:"column:is_active;type:boolean;not null;default:true"`
	PermissionBindings []*PermissionGroupPermission `gorm:"foreignKey:PermissionGroupID;-:migration"`
}

func (PermissionGroup) TableName() string {
	return "permission_groups"
}
