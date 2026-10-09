package models

import "time"

type Permission struct {
	ID          int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Code        string     `gorm:"column:code;type:varchar(255);not null;uniqueIndex:permission_code"`
	Description string     `gorm:"column:description;type:varchar(255);not null"`
	IsActive    *bool      `gorm:"column:is_active;type:boolean;not null;default:true"`
}

func (Permission) TableName() string {
	return "permissions"
}
