package account

import "time"

type Account struct {
	ID           int
	Name         string
	Email        string
	Password     string
	WechatOpenID string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
