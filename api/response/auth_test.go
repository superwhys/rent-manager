package response_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"

	"github.com/superwhys/rent-manager/api/response"
)

func TestHandleAuthErrorEmailFailures(t *testing.T) {
	for _, tt := range []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{name: "unregistered", err: authkit.ErrRegistrationDenied, status: http.StatusForbidden, message: "该邮箱尚未注册，请先注册"},
		{name: "already registered", err: authkit.ErrConflict, status: http.StatusConflict, message: "该邮箱已注册，请直接登录"},
		{name: "resend too soon", err: authkit.ErrResendTooSoon, status: http.StatusTooManyRequests, message: "验证码发送过于频繁，请稍后重试"},
		{name: "wrong code", err: authkit.ErrChallengeMismatch, status: http.StatusBadRequest, message: "验证码错误，请重新输入"},
		{name: "expired code", err: authkit.ErrChallengeInvalid, status: http.StatusBadRequest, message: "验证码已失效，请重新获取"},
		{name: "replaced code", err: authkit.ErrChallengeUpdated, status: http.StatusConflict, message: "验证码已更新，请使用最新验证码"},
		{name: "mail not configured", err: authkit.ErrEmailUnavailable, status: http.StatusServiceUnavailable, message: "邮箱验证码服务尚未配置"},
		{name: "mail failed", err: authkit.ErrMailFailed, status: http.StatusBadGateway, message: "验证码发送失败，请稍后重试"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/email/code", nil)
			if !response.HandleAuthError(ctx, fmt.Errorf("wrapped: %w", tt.err)) {
				t.Fatal("error was not handled")
			}
			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tt.status || body.Code == 0 || body.Message != tt.message {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
