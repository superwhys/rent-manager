package config

import (
	"errors"
	"time"
)

type JwtConfig struct {
	JwtSecret         string
	JwtTimeout        time.Duration
	JwtRefreshTimeout time.Duration
}

func (c *JwtConfig) Validate() error {
	if c.JwtSecret == "" {
		return errors.New("JWT 密钥不能为空")
	}
	if c.JwtTimeout <= 0 {
		return errors.New("JWT 过期时间不能为空")
	}
	if c.JwtRefreshTimeout <= 0 {
		c.JwtRefreshTimeout = c.JwtTimeout * 24
	}
	return nil
}

type WechatConfig struct {
	AppID             string
	SecretID          string
	Jscode2SessionApi string
}

func (c *WechatConfig) Validate() error {
	if c.AppID == "" {
		return errors.New("微信 appid 不能为空")
	}
	if c.SecretID == "" {
		return errors.New("微信 secretid 不能为空")
	}
	if c.Jscode2SessionApi == "" {
		return errors.New("微信 jscode2sessionapi 不能为空")
	}
	return nil
}

type AdminConfig struct {
	SuperAdminAccountIDs []int
}

func (c *AdminConfig) IsSuperAdmin(accountID int) bool {
	if accountID <= 0 {
		return false
	}
	for _, id := range c.SuperAdminAccountIDs {
		if id == accountID {
			return true
		}
	}
	return false
}
