package models

import "time"

type AccountSetting struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	Key       string     `gorm:"column:key;type:varchar(64);not null;uniqueIndex:accountsetting_account_id_key,priority:2"`
	Value     string     `gorm:"column:value;type:varchar(255);not null"`
	AccountID string     `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:accountsetting_account_id_key,priority:1"`
}

func (AccountSetting) TableName() string {
	return "account_settings"
}
