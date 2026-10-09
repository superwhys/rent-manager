package models

import "time"

type Account struct {
	ID           int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Name         string     `gorm:"column:name;type:varchar(255);not null"`
	Email        string     `gorm:"column:email;type:varchar(255);not null"`
	Password     string     `gorm:"column:password;type:varchar(255);not null"`
	WechatOpenid string     `gorm:"column:wechat_openid;type:varchar(255);not null;uniqueIndex:account_wechat_openid"`
	IsActive     *bool      `gorm:"column:is_active;type:boolean;not null;default:true"`
}

func (Account) TableName() string {
	return "accounts"
}
