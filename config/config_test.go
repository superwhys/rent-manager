package config

import (
	"strings"
	"testing"
	"time"
)

func TestAuthConfigValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		config AuthConfig
		valid  bool
	}{
		{name: "未配置邮件允许已有账号密码登录", valid: true},
		{name: "负会话有效期", config: AuthConfig{SessionTTL: -time.Hour}},
		{name: "指定会话有效期", config: AuthConfig{SessionTTL: time.Hour}, valid: true},
		{name: "邮件配置不完整", config: AuthConfig{SMTP: SMTPConfig{Host: "smtp.example.com"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.Validate(); (err == nil) != test.valid {
				t.Fatalf("Validate() = %v, valid = %v", err, test.valid)
			}
		})
	}
}

func TestSMTPConfigValidation(t *testing.T) {
	complete := SMTPConfig{
		Host: "smtp.example.com", Port: 587, From: "notice@example.com",
		Username: "notice@example.com", Password: "mail-password", Timeout: 10 * time.Second,
	}
	for _, test := range []struct {
		name  string
		edit  func(*SMTPConfig)
		valid bool
	}{
		{name: "完整配置", valid: true},
		{name: "默认超时", edit: func(c *SMTPConfig) { c.Timeout = 0 }, valid: true},
		{name: "无认证邮件网关", edit: func(c *SMTPConfig) { c.Username, c.Password = "", "" }, valid: true},
		{name: "只填用户名", edit: func(c *SMTPConfig) { c.Password = "" }},
		{name: "只填密码", edit: func(c *SMTPConfig) { c.Username = "" }},
		{name: "缺少服务地址", edit: func(c *SMTPConfig) { c.Host = "" }},
		{name: "服务地址空白", edit: func(c *SMTPConfig) { c.Host = " smtp.example.com" }},
		{name: "地址包含端口", edit: func(c *SMTPConfig) { c.Host = "smtp.example.com:587" }},
		{name: "缺少端口", edit: func(c *SMTPConfig) { c.Port = 0 }},
		{name: "无效端口", edit: func(c *SMTPConfig) { c.Port = 65536 }},
		{name: "缺少发件邮箱", edit: func(c *SMTPConfig) { c.From = "" }},
		{name: "无效发件邮箱", edit: func(c *SMTPConfig) { c.From = "invalid" }},
		{name: "发件邮箱不允许显示名称", edit: func(c *SMTPConfig) { c.From = "Name <notice@example.com>" }},
		{name: "发件邮箱注入", edit: func(c *SMTPConfig) { c.From = "notice@example.com\r\nBcc: other@example.com" }},
		{name: "发件名称注入", edit: func(c *SMTPConfig) { c.FromName = "租住\r\nBcc: other@example.com" }},
		{name: "负超时", edit: func(c *SMTPConfig) { c.Timeout = -time.Second }},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := complete
			if test.edit != nil {
				test.edit(&c)
			}
			if err := c.Validate(); (err == nil) != test.valid {
				t.Fatalf("Validate() = %v, valid = %v", err, test.valid)
			}
		})
	}
	if (SMTPConfig{}).Configured() {
		t.Fatal("空配置不能启用邮件发送")
	}
	if err := (SMTPConfig{}).Validate(); err != nil {
		t.Fatalf("未配置邮件应允许启动: %v", err)
	}
}

func TestSuperAdminUsesExactAuthkitID(t *testing.T) {
	id := strings.Repeat("ab", 32)
	config := AdminConfig{SuperAdminAccountIDs: []string{id, ""}}
	if !config.IsSuperAdmin(id) {
		t.Fatal("authkit 管理员 ID 应匹配")
	}
	for _, other := range []string{"", "1", strings.ToUpper(id), strings.Repeat("c", 64)} {
		if config.IsSuperAdmin(other) {
			t.Fatalf("不应将 %q 识别为管理员", other)
		}
	}
}
