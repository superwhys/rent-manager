package config

import (
	"errors"
	"strings"
	"time"

	"github.com/miebyte/authkit/wechat"
)

type AuthConfig struct {
	Wechat     wechat.Config `json:"wechat"`
	SessionTTL time.Duration `json:"session_ttl"`
}

func (c *AuthConfig) Validate() error {
	if err := c.Wechat.Validate(); err != nil {
		return errors.New("微信 app_id 与 secret 必须同时配置且不能包含首尾空白")
	}
	if !c.Wechat.Enabled() {
		return errors.New("微信 app_id 与 secret 不能为空")
	}
	if c.SessionTTL < 0 {
		return errors.New("会话有效期不能为负数")
	}
	return nil
}

type AdminConfig struct {
	SuperAdminAccountIDs []string
}

func (c *AdminConfig) IsSuperAdmin(accountID string) bool {
	if strings.TrimSpace(accountID) == "" {
		return false
	}
	for _, id := range c.SuperAdminAccountIDs {
		if id == accountID {
			return true
		}
	}
	return false
}
