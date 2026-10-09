package config

import (
	"strings"
	"testing"
	"time"

	"github.com/miebyte/authkit/wechat"
)

func TestAuthConfigValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		config AuthConfig
		valid  bool
	}{
		{name: "未配置微信"},
		{name: "缺少密钥", config: AuthConfig{Wechat: wechat.Config{AppID: "app"}}},
		{name: "凭证空白", config: AuthConfig{Wechat: wechat.Config{AppID: " app", Secret: "secret"}}},
		{name: "负会话有效期", config: AuthConfig{Wechat: wechat.Config{AppID: "app", Secret: "secret"}, SessionTTL: -time.Hour}},
		{name: "使用 authkit 默认有效期", config: AuthConfig{Wechat: wechat.Config{AppID: "app", Secret: "secret"}}, valid: true},
		{name: "指定会话有效期", config: AuthConfig{Wechat: wechat.Config{AppID: "app", Secret: "secret"}, SessionTTL: time.Hour}, valid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.Validate(); (err == nil) != test.valid {
				t.Fatalf("Validate() = %v, valid = %v", err, test.valid)
			}
		})
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
