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
	AccountID        int                `gorm:"column:account_id;type:bigint;not null;uniqueIndex:role_account_id_name,priority:1"`
	Account          *Account           `gorm:"foreignKey:AccountID;constraint:roles_accounts_roles,OnDelete:NO ACTION"`
	PermissionGroups []*PermissionGroup `gorm:"many2many:role_permission_groups;joinForeignKey:RoleID;joinReferences:PermissionGroupID;-:migration"`
}

func (Role) TableName() string {
	return "roles"
}
