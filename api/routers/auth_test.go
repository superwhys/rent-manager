package routers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/routers"
	"github.com/superwhys/rent-manager/config"
	appcore "github.com/superwhys/rent-manager/internal/app"
	appauth "github.com/superwhys/rent-manager/internal/app/auth"
)

type authServiceStub struct {
	sendCode      func(context.Context, authkit.SendCodeInput) error
	loginEmail    func(context.Context, authkit.EmailLoginInput) (*authkit.LoginResult, error)
	register      func(context.Context, appauth.RegisterRequest, string) (*authkit.LoginResult, error)
	loginPassword func(context.Context, authkit.PasswordLoginInput) (*authkit.LoginResult, error)
	authenticate  func(context.Context, string) (*authkit.AuthenticatedSession, error)
	logout        func(context.Context, string) error
}

func (s authServiceStub) LoginPassword(
	ctx context.Context,
	input authkit.PasswordLoginInput,
) (*authkit.LoginResult, error) {
	return s.loginPassword(ctx, input)
}

func (s authServiceStub) SendCode(ctx context.Context, input authkit.SendCodeInput) error {
	return s.sendCode(ctx, input)
}

func (s authServiceStub) LoginEmail(
	ctx context.Context,
	input authkit.EmailLoginInput,
) (*authkit.LoginResult, error) {
	return s.loginEmail(ctx, input)
}

func (s authServiceStub) Register(
	ctx context.Context,
	input appauth.RegisterRequest,
	ip string,
) (*authkit.LoginResult, error) {
	return s.register(ctx, input, ip)
}

func (s authServiceStub) AuthenticateSession(
	ctx context.Context,
	token string,
) (*authkit.AuthenticatedSession, error) {
	return s.authenticate(ctx, token)
}

func (s authServiceStub) Logout(ctx context.Context, token string) error {
	return s.logout(ctx, token)
}

func newAuthHandler(service authServiceStub) http.Handler {
	return newAuthHandlerWithAdminConfig(service, nil)
}

func newAuthHandlerWithAdminConfig(service authServiceStub, adminConfig *config.AdminConfig) http.Handler {
	return ginutils.NewServerHandler(ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/auth"),
		routers.AuthRouter(service, adminConfig),
	))
}

func TestAuthRouterSendEmailCodeReturnsOnlyTiming(t *testing.T) {
	calls := 0
	handler := newAuthHandler(authServiceStub{
		sendCode: func(_ context.Context, input authkit.SendCodeInput) error {
			calls++
			if input.Email != "tenant@example.com" || input.IP != "192.0.2.8" {
				t.Errorf("send code input = %+v", input)
			}
			return nil
		},
	})
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/email/code?email=override@example.com",
		strings.NewReader(`{"email":"tenant@example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Email", "header@example.com")
	req.RemoteAddr = "192.0.2.8:4567"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var body struct {
		Code int            `json:"code"`
		Data map[string]int `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || body.Code != 0 || calls != 1 || len(body.Data) != 2 ||
		body.Data["resend_after_seconds"] != 60 ||
		body.Data["expires_in_seconds"] != 600 {
		t.Fatalf(
			"unexpected code response: status=%d body=%s",
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestAuthRouterEmailRegistrationAndCodeLogin(t *testing.T) {
	result := &authkit.LoginResult{
		Account: authkit.Account{ID: strings.Repeat("a", 64), Email: "tenant@example.com"},
		Token:   strings.Repeat("b", 64), Expires: time.Now().Add(time.Hour),
	}
	for _, tt := range []struct {
		path string
		body string
	}{
		{path: "/auth/register", body: `{"email":"tenant@example.com","code":"123456","password":"  actual-password  "}`},
		{path: "/auth/email-login", body: `{"email":"tenant@example.com","code":"123456"}`},
	} {
		t.Run(tt.path, func(t *testing.T) {
			calls := 0
			handler := newAuthHandler(authServiceStub{
				register: func(_ context.Context, input appauth.RegisterRequest, ip string) (*authkit.LoginResult, error) {
					calls++
					if input.Email != "tenant@example.com" || input.Code != "123456" ||
						input.Password != "  actual-password  " ||
						ip != "192.0.2.8" {
						t.Error("registration input differs from JSON body or client IP")
					}
					return result, nil
				},
				loginEmail: func(_ context.Context, input authkit.EmailLoginInput) (*authkit.LoginResult, error) {
					calls++
					if input.Email != "tenant@example.com" || input.Code != "123456" {
						t.Error("code login input differs from JSON body")
					}
					return result, nil
				},
			})
			req := httptest.NewRequest(
				http.MethodPost,
				tt.path+"?email=override@example.com&code=654321&password=override",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Email", "header@example.com")
			req.Header.Set("Code", "654321")
			req.Header.Set("Password", "header-password")
			req.RemoteAddr = "192.0.2.8:4567"
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			var body struct {
				Code int                   `json:"code"`
				Data routers.LoginResponse `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != http.StatusOK || body.Code != 0 || calls != 1 ||
				body.Data.Account != result.Account ||
				body.Data.AccessToken != result.Token {
				t.Fatalf(
					"unexpected response: status=%d body=%s",
					recorder.Code,
					recorder.Body.String(),
				)
			}
		})
	}
}

func TestAuthRouterRejectsInvalidRegistrationAndCode(t *testing.T) {
	for _, tt := range []struct {
		name string
		path string
		body string
	}{
		{name: "weak password", path: "/auth/register", body: `{"email":"tenant@example.com","code":"123456","password":"short"}`},
		{name: "missing code", path: "/auth/register", body: `{"email":"tenant@example.com","password":"valid-password"}`},
		{name: "invalid code", path: "/auth/email-login", body: `{"email":"tenant@example.com","code":"abcdef"}`},
		{name: "short code", path: "/auth/email-login", body: `{"email":"tenant@example.com","code":"123"}`},
		{name: "signed code", path: "/auth/email-login", body: `{"email":"tenant@example.com","code":"+12345"}`},
		{name: "empty email", path: "/auth/email/code", body: `{"email":""}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := newAuthHandler(authServiceStub{})
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Code == 0 {
				t.Fatalf("invalid input accepted: body=%s err=%v", recorder.Body.String(), err)
			}
			if recorder.Header().Get("Cache-Control") != "no-store" {
				t.Error("authentication errors must not be cached")
			}
		})
	}
}

func TestAuthRouterPasswordLoginReturnsAuthkitSession(t *testing.T) {
	result := &authkit.LoginResult{
		Account: authkit.Account{
			ID:       strings.Repeat("a", 64),
			Username: "租户",
			Email:    "tenant@example.com",
		},
		Token:   strings.Repeat("b", 64),
		Expires: time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC),
	}
	type requestKey struct{}
	ctx := context.WithValue(context.Background(), requestKey{}, "request-context")
	loginCalls := 0
	handler := newAuthHandler(authServiceStub{
		loginPassword: func(gotCtx context.Context, input authkit.PasswordLoginInput) (*authkit.LoginResult, error) {
			loginCalls++
			want := authkit.PasswordLoginInput{
				Identifier: "tenant@example.com",
				Password:   "  actual-password  ",
				IP:         "192.0.2.8",
			}
			if input != want {
				t.Error(
					"email normalization, password preservation, or client IP differs from expected input",
				)
			}
			if gotCtx.Value(requestKey{}) != "request-context" {
				t.Error("request context was not propagated")
			}
			return result, nil
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/password-login?email=override@example.com&password=override", strings.NewReader(`{"email":" TENANT@EXAMPLE.COM ","password":"  actual-password  "}`)).
		WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Email", "header@example.com")
	req.Header.Set("Password", "header-password")
	req.RemoteAddr = "192.0.2.8:4567"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || loginCalls != 1 {
		t.Fatalf(
			"status = %d, login calls = %d, body = %s",
			recorder.Code,
			loginCalls,
			recorder.Body.String(),
		)
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Error("login response must not be cached")
	}
	var body struct {
		Code int                   `json:"code"`
		Data routers.LoginResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Account != result.Account ||
		body.Data.AccessToken != result.Token ||
		!body.Data.ExpiresAt.Equal(result.Expires) ||
		body.Data.Created {
		t.Fatalf("unexpected email login response: %s", recorder.Body.String())
	}
}

func TestAuthRouterSuperAdminIdentityMatchesLoginAndSession(t *testing.T) {
	adminID := strings.Repeat("a", 64)
	for _, tt := range []struct {
		name        string
		accountID   string
		adminConfig *config.AdminConfig
		want        bool
	}{
		{
			name: "super admin", accountID: adminID,
			adminConfig: &config.AdminConfig{SuperAdminAccountIDs: []string{adminID}},
			want:        true,
		},
		{
			name: "ordinary account", accountID: strings.Repeat("c", 64),
			adminConfig: &config.AdminConfig{SuperAdminAccountIDs: []string{adminID}},
		},
		{name: "empty configuration", accountID: adminID, adminConfig: &config.AdminConfig{}},
		{name: "nil configuration", accountID: adminID},
		{
			name: "case differs", accountID: adminID,
			adminConfig: &config.AdminConfig{SuperAdminAccountIDs: []string{strings.ToUpper(adminID)}},
		},
		{
			name: "whitespace differs", accountID: adminID,
			adminConfig: &config.AdminConfig{SuperAdminAccountIDs: []string{" " + adminID + " "}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := &authkit.LoginResult{
				Account: authkit.Account{ID: tt.accountID, Email: "tenant@example.com"},
				Token:   strings.Repeat("b", 64), Expires: time.Now().Add(time.Hour),
			}
			for _, route := range []struct {
				path string
				body string
			}{
				{path: "/auth/register", body: `{"email":"tenant@example.com","code":"123456","password":"valid-password","is_super_admin":true}`},
				{path: "/auth/email-login", body: `{"email":"tenant@example.com","code":"123456","is_super_admin":true}`},
				{path: "/auth/password-login", body: `{"email":"tenant@example.com","password":"valid-password","is_super_admin":true}`},
			} {
				t.Run(route.path, func(t *testing.T) {
					handler := newAuthHandlerWithAdminConfig(authServiceStub{
						register: func(context.Context, appauth.RegisterRequest, string) (*authkit.LoginResult, error) {
							return result, nil
						},
						loginEmail: func(context.Context, authkit.EmailLoginInput) (*authkit.LoginResult, error) {
							return result, nil
						},
						loginPassword: func(context.Context, authkit.PasswordLoginInput) (*authkit.LoginResult, error) {
							return result, nil
						},
						authenticate: func(_ context.Context, token string) (*authkit.AuthenticatedSession, error) {
							if token != result.Token {
								t.Errorf("session token = %q, want login token %q", token, result.Token)
							}
							return &authkit.AuthenticatedSession{Account: result.Account, Expires: result.Expires}, nil
						},
					}, tt.adminConfig)
					token := ""
					for _, request := range []struct {
						method string
						path   string
						body   string
					}{
						{method: http.MethodPost, path: route.path, body: route.body},
						{method: http.MethodGet, path: "/auth/session"},
					} {
						req := httptest.NewRequest(request.method, request.path, strings.NewReader(request.body))
						req.Header.Set("Content-Type", "application/json")
						if request.method == http.MethodGet {
							req.Header.Set("Authorization", "Bearer "+token)
						}
						recorder := httptest.NewRecorder()
						handler.ServeHTTP(recorder, req)
						var body struct {
							Code int `json:"code"`
							Data struct {
								Account      authkit.Account `json:"account"`
								IsSuperAdmin *bool           `json:"is_super_admin"`
								AccessToken  string          `json:"access_token"`
							} `json:"data"`
						}
						if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
							t.Fatal(err)
						}
						if recorder.Code != http.StatusOK || body.Code != 0 || body.Data.Account != result.Account ||
							body.Data.IsSuperAdmin == nil || *body.Data.IsSuperAdmin != tt.want {
							t.Fatalf("unexpected %s response: status=%d body=%s", request.path, recorder.Code, recorder.Body.String())
						}
						if request.method == http.MethodPost {
							token = body.Data.AccessToken
							if token != result.Token {
								t.Fatalf("login token = %q, want %q", token, result.Token)
							}
						}
						if recorder.Header().Get("Cache-Control") != "no-store" {
							t.Errorf("%s response must not be cached", request.path)
						}
					}
				})
			}
		})
	}
}

func TestAuthRouterPasswordLoginRejectsInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "missing email", body: `{"password":"private-test-password"}`},
		{name: "missing password", body: `{"email":"tenant@example.com"}`},
		{name: "empty password", body: `{"email":"tenant@example.com","password":""}`},
		{name: "invalid email", body: `{"email":"tenant","password":"private-test-password"}`},
		{name: "display name email", body: `{"email":"Tenant <tenant@example.com>","password":"private-test-password"}`},
		{name: "long password", body: `{"email":"tenant@example.com","password":"` + strings.Repeat("x", 129) + `"}`},
		{name: "malformed json", body: `{"email":"tenant@example.com","password":"private-test-password"`},
		{name: "non-string password", body: `{"email":"tenant@example.com","password":123}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			loginCalls := 0
			handler := newAuthHandler(authServiceStub{
				loginPassword: func(context.Context, authkit.PasswordLoginInput) (*authkit.LoginResult, error) {
					loginCalls++
					return nil, authkit.ErrInvalidCredentials
				},
			})
			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/password-login?email=query@example.com&password=query-password",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Email", "header@example.com")
			req.Header.Set("Password", "header-password")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid error response: %s", recorder.Body.String())
			}
			if body.Code == 0 || loginCalls != 0 {
				t.Fatalf(
					"code = %d, login calls = %d, body = %s",
					body.Code,
					loginCalls,
					recorder.Body.String(),
				)
			}
			if strings.Contains(recorder.Body.String(), "private-test-password") {
				t.Error("error response exposed the password")
			}
		})
	}
}

func TestAuthRouterPasswordLoginMapsAuthErrors(t *testing.T) {
	for _, tt := range []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{name: "invalid credentials", err: authkit.ErrInvalidCredentials, status: http.StatusUnauthorized, message: "邮箱或密码错误"},
		{name: "disabled account", err: authkit.ErrBlacklisted, status: http.StatusForbidden, message: "该账号暂不允许登录"},
		{name: "rate limited", err: authkit.ErrTooManyRequests, status: http.StatusTooManyRequests, message: "请求过于频繁"},
		{name: "invalid input", err: authkit.ErrInvalidInput, status: http.StatusBadRequest, message: "登录参数无效"},
		{name: "invalid email", err: authkit.ErrInvalidEmail, status: http.StatusBadRequest, message: "邮箱格式不正确"},
		{name: "internal error", err: errors.New("storage unavailable"), status: http.StatusInternalServerError, message: "系统内部错误"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := newAuthHandler(authServiceStub{
				loginPassword: func(context.Context, authkit.PasswordLoginInput) (*authkit.LoginResult, error) {
					return nil, tt.err
				},
			})
			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/password-login",
				strings.NewReader(
					`{"email":"tenant@example.com","password":"private-test-password"}`,
				),
			)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid error response: %s", recorder.Body.String())
			}
			if recorder.Code != tt.status || body.Code == 0 || body.Message != tt.message {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestAuthRouterSessionAndLogout(t *testing.T) {
	token := strings.Repeat("b", 64)
	session := &authkit.AuthenticatedSession{
		Account: authkit.Account{ID: strings.Repeat("a", 64), Username: "租户"},
		Method:  authkit.MethodEmail,
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
		t.Fatalf(
			"session status = %d, authentication calls = %d, body = %s",
			recorder.Code,
			authCalls,
			recorder.Body.String(),
		)
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
	if body.Code != 0 || body.Data.Account != session.Account ||
		!body.Data.ExpiresAt.Equal(session.Expires) {
		t.Fatalf("unexpected session response: %s", recorder.Body.String())
	}
	req = httptest.NewRequest(http.MethodDelete, "/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK || authCalls != 2 || logoutCalls != 1 {
		t.Fatalf(
			"logout status = %d, authentication calls = %d, logout calls = %d, body = %s",
			recorder.Code,
			authCalls,
			logoutCalls,
			recorder.Body.String(),
		)
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
	for _, path := range []string{"/auth/refresh-token", "/auth/wechat-login"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Fatalf(
				"removed route %s status = %d, want 404; body = %s",
				path,
				recorder.Code,
				recorder.Body.String(),
			)
		}
	}
}
