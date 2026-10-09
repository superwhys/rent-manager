package models

import "time"

type AccountSetting struct {
	ID        int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Key       string     `gorm:"column:key;type:varchar(64);not null;uniqueIndex:accountsetting_account_id_key,priority:2"`
	Value     string     `gorm:"column:value;type:varchar(255);not null"`
	AccountID int        `gorm:"column:account_id;type:bigint;not null;uniqueIndex:accountsetting_account_id_key,priority:1"`
	Account   *Account   `gorm:"foreignKey:AccountID;constraint:account_settings_accounts_account_settings,OnDelete:NO ACTION"`
}

func (AccountSetting) TableName() string {
	return "account_settings"
}
