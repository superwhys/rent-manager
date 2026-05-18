package account

import (
	"fmt"
	"strings"
	"time"
)

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

func NewWechatAccount(openID string) (*Account, error) {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return nil, ErrWechatOpenIDRequired
	}
	return &Account{
		Name:         fmt.Sprintf("wechat_%s", openID),
		Email:        fmt.Sprintf("%s@wechat.local", openID),
		Password:     openID,
		WechatOpenID: openID,
		IsActive:     true,
	}, nil
}
