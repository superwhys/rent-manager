package routers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/routers"
	appcore "github.com/superwhys/rent-manager/internal/app"
)

type authServiceStub struct {
	login        func(context.Context, string) (*authkit.LoginResult, error)
	authenticate func(context.Context, string) (*authkit.AuthenticatedSession, error)
	logout       func(context.Context, string) error
}

func (s authServiceStub) LoginWechat(ctx context.Context, code string) (*authkit.LoginResult, error) {
	return s.login(ctx, code)
}

func (s authServiceStub) AuthenticateSession(ctx context.Context, token string) (*authkit.AuthenticatedSession, error) {
	return s.authenticate(ctx, token)
}

func (s authServiceStub) Logout(ctx context.Context, token string) error {
	return s.logout(ctx, token)
}

func newAuthHandler(service authServiceStub) http.Handler {
	return ginutils.NewServerHandler(ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/auth"),
		routers.AuthRouter(service),
	))
}

func TestAuthRouterWechatLoginReturnsAuthkitSession(t *testing.T) {
	result := &authkit.LoginResult{
		Account: authkit.Account{ID: strings.Repeat("a", 64), Username: "租户", Email: "tenant@example.com"},
		Token:   strings.Repeat("b", 64),
		Expires: time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
		Created: true,
	}
	loginCalls := 0
	handler := newAuthHandler(authServiceStub{
		login: func(_ context.Context, code string) (*authkit.LoginResult, error) {
			loginCalls++
			if code != "wechat-code" {
				t.Errorf("code = %q, want wechat-code", code)
			}
			return result, nil
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/wechat-login", strings.NewReader(`{"code":"wechat-code"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || loginCalls != 1 {
		t.Fatalf("status = %d, login calls = %d, body = %s", recorder.Code, loginCalls, recorder.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			Account     authkit.Account `json:"account"`
			AccessToken string          `json:"access_token"`
			ExpiresAt   time.Time       `json:"expires_at"`
			Created     bool            `json:"created"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Account != result.Account || body.Data.AccessToken != result.Token || !body.Data.ExpiresAt.Equal(result.Expires) || !body.Data.Created {
		t.Fatalf("unexpected login response: %s", recorder.Body.String())
	}
	var raw struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, exists := raw.Data["refresh_token"]; exists {
		t.Error("authkit login must not issue a legacy refresh token")
	}
	var account map[string]json.RawMessage
	if err := json.Unmarshal(raw.Data["account"], &account); err != nil {
		t.Fatal(err)
	}
	if _, exists := account["ID"]; !exists {
		t.Errorf("account must use authkit.Account directly, got %s", raw.Data["account"])
	}
}

func TestAuthRouterSessionAndLogout(t *testing.T) {
	token := strings.Repeat("b", 64)
	session := &authkit.AuthenticatedSession{
		Account: authkit.Account{ID: strings.Repeat("a", 64), Username: "租户"},
		Method:  authkit.MethodWechat,
		Expires: time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
	}
	authCalls, logoutCalls := 0, 0
	handler := newAuthHandler(authServiceStub{
		authenticate: func(_ context.Context, gotToken string) (*authkit.AuthenticatedSession, error) {
			authCalls++
			if gotToken != token {
				t.Errorf("authentication token = %q, want %q", gotToken, token)
			}
			return session, nil
		},
		logout: func(ctx context.Context, gotToken string) error {
			logoutCalls++
			if gotToken != token {
				t.Errorf("logout token = %q, want %q", gotToken, token)
			}
			accountID, err := appcore.AccountIDFromContext(ctx)
			if err != nil || accountID != session.Account.ID {
				t.Errorf("logout account ID = %q, error = %v", accountID, err)
			}
			return nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || authCalls != 1 {
		t.Fatalf("session status = %d, authentication calls = %d, body = %s", recorder.Code, authCalls, recorder.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			Account   authkit.Account `json:"account"`
			ExpiresAt time.Time       `json:"expires_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Account != session.Account || !body.Data.ExpiresAt.Equal(session.Expires) {
		t.Fatalf("unexpected session response: %s", recorder.Body.String())
	}
	req = httptest.NewRequest(http.MethodDelete, "/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || authCalls != 2 || logoutCalls != 1 {
		t.Fatalf("logout status = %d, authentication calls = %d, logout calls = %d, body = %s", recorder.Code, authCalls, logoutCalls, recorder.Body.String())
	}
}

func TestAuthRouterRequiresSessionAndRemovesRefreshToken(t *testing.T) {
	handler := newAuthHandler(authServiceStub{})
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(method, "/auth/session", nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body = %s", recorder.Code, recorder.Body.String())
			}
		})
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/auth/refresh-token", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("legacy refresh status = %d, want 404; body = %s", recorder.Code, recorder.Body.String())
	}
}
