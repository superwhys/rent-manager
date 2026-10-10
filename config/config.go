package config

import (
	"errors"
	"net"
	"net/mail"
	"slices"
	"strings"
	"time"
)

type AuthConfig struct {
	SMTP       SMTPConfig    `json:"smtp"`
	SessionTTL time.Duration `json:"session_ttl"`
}

func (c *AuthConfig) Validate() error {
	if c.SessionTTL < 0 {
		return errors.New("会话有效期不能为负数")
	}
	return c.SMTP.Validate()
}

type SMTPConfig struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	From     string        `json:"from"`
	FromName string        `json:"from_name"`
	Timeout  time.Duration `json:"timeout"`
}

func (c SMTPConfig) Configured() bool {
	return c.Host != "" || c.Port != 0 || c.Username != "" || c.Password != "" ||
		c.From != "" || c.FromName != "" || c.Timeout != 0
}

func (c SMTPConfig) Validate() error {
	if !c.Configured() {
		return nil
	}
	if c.Host == "" || strings.ContainsAny(c.Host, " \t\r\n/[]") ||
		(strings.Contains(c.Host, ":") && net.ParseIP(c.Host) == nil) {
		return errors.New("邮件服务器地址不能为空，且不能包含端口、路径或空白")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("邮件服务器端口必须在 1 到 65535 之间")
	}
	if (c.Username == "") != (c.Password == "") || strings.TrimSpace(c.Username) != c.Username {
		return errors.New("邮件用户名与密码必须同时配置，用户名不能包含首尾空白")
	}
	address, err := mail.ParseAddress(c.From)
	if err != nil || address.Address != c.From || strings.ContainsAny(c.From, "\r\n") {
		return errors.New("发件邮箱格式不正确，请单独填写邮箱地址")
	}
	if strings.ContainsAny(c.FromName, "\r\n") {
		return errors.New("发件人名称不能包含换行")
	}
	if c.Timeout < 0 {
		return errors.New("邮件发送超时不能为负数")
	}
	return nil
}

type AdminConfig struct {
	SuperAdminAccountIDs []string `json:"super_admin_account_ids"`
}

func (c *AdminConfig) IsSuperAdmin(accountID string) bool {
	if strings.TrimSpace(accountID) == "" {
		return false
	}

	if slices.Contains(c.SuperAdminAccountIDs, accountID) {
		return true
	}

	return false
}
