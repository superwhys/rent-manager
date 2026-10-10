package middlewares_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/authkit"

	"github.com/superwhys/rent-manager/api/middlewares"
	appcore "github.com/superwhys/rent-manager/internal/app"
)

type sessionAuthenticatorFunc func(context.Context, string) (*authkit.AuthenticatedSession, error)

func (f sessionAuthenticatorFunc) AuthenticateSession(
	ctx context.Context,
	token string,
) (*authkit.AuthenticatedSession, error) {
	return f(ctx, token)
}

func TestTokenVerifyMiddlewarePropagatesAuthkitIdentity(t *testing.T) {
	token := strings.Repeat("b", 64)
	session := &authkit.AuthenticatedSession{
		Account: authkit.Account{ID: strings.Repeat("a", 64), Username: "租户"},
		Method:  authkit.MethodEmail,
		Expires: time.Now().Add(time.Hour),
	}
	type requestContextKey struct{}
	calls := 0
	authenticator := sessionAuthenticatorFunc(
		func(ctx context.Context, gotToken string) (*authkit.AuthenticatedSession, error) {
			calls++
			if gotToken != token {
				t.Errorf("token = %q, want %q", gotToken, token)
			}
			if ctx.Value(requestContextKey{}) != "request context" {
				t.Error("authenticator did not receive the request context")
			}
			return session, nil
		},
	)
	engine := gin.New()
	engine.GET(
		"/protected",
		middlewares.TokenVerifyMiddleware(authenticator),
		func(ctx *gin.Context) {
			accountID, err := appcore.AccountIDFromContext(ctx.Request.Context())
			if err != nil || accountID != session.Account.ID {
				t.Errorf("account ID = %q, error = %v, want %q", accountID, err, session.Account.ID)
			}
			if got := middlewares.AuthenticatedSessionFromContext(ctx); got != session {
				t.Errorf("session = %#v, want %#v", got, session)
			}
			ctx.Status(http.StatusNoContent)
		},
	)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestContextKey{}, "request context"))
	req.Header.Set("Authorization", "bEaReR "+token)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent || calls != 1 {
		t.Fatalf(
			"status = %d, authentication calls = %d, body = %s",
			recorder.Code,
			calls,
			recorder.Body.String(),
		)
	}
}

func TestTokenVerifyMiddlewareRejectsInvalidSessions(t *testing.T) {
	token := strings.Repeat("b", 64)
	session := &authkit.AuthenticatedSession{Account: authkit.Account{ID: strings.Repeat("a", 64)}}
	tests := []struct {
		name       string
		header     string
		session    *authkit.AuthenticatedSession
		err        error
		wantStatus int
		wantCalls  int
	}{
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "raw token", header: token, wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic " + token, wantStatus: http.StatusUnauthorized},
		{name: "missing token", header: "Bearer ", wantStatus: http.StatusUnauthorized},
		{
			name:       "extra token",
			header:     "Bearer " + token + " extra",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired session",
			header:     "Bearer " + token,
			err:        authkit.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
		{
			name:       "revoked session",
			header:     "Bearer " + token,
			err:        fmt.Errorf("revoked: %w", authkit.ErrUnauthorized),
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
		{
			name:       "blacklisted account",
			header:     "Bearer " + token,
			err:        fmt.Errorf("blocked: %w", authkit.ErrBlacklisted),
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
		{
			name:       "store failure",
			header:     "Bearer " + token,
			err:        errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCalls:  1,
		},
		{
			name:   "impersonation",
			header: "Bearer " + token,
			session: &authkit.AuthenticatedSession{
				Account: session.Account,
				Actor:   &authkit.Account{ID: strings.Repeat("c", 64)},
			},
			wantStatus: http.StatusUnauthorized,
			wantCalls:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			authenticator := sessionAuthenticatorFunc(
				func(context.Context, string) (*authkit.AuthenticatedSession, error) {
					calls++
					if tt.session != nil {
						return tt.session, tt.err
					}
					return session, tt.err
				},
			)
			engine := gin.New()
			engine.GET(
				"/protected",
				middlewares.TokenVerifyMiddleware(authenticator),
				func(ctx *gin.Context) {
					t.Error("unauthorized request reached the protected handler")
					ctx.Status(http.StatusNoContent)
				},
			)
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatus || calls != tt.wantCalls {
				t.Fatalf(
					"status = %d, calls = %d, want status = %d, calls = %d; body = %s",
					recorder.Code,
					calls,
					tt.wantStatus,
					tt.wantCalls,
					recorder.Body.String(),
				)
			}
		})
	}
}
