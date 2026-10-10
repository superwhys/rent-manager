package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"

	"github.com/superwhys/rent-manager/api"
	appadmin "github.com/superwhys/rent-manager/internal/app/admin"
)

func TestSetupRouterDoesNotTrustForwardedClientIP(t *testing.T) {
	engine := api.SetupRouter(&limiter.Limiter{}, nil, &appadmin.Services{})
	for _, test := range []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{name: "direct", remoteAddr: "192.0.2.8:4567", want: "192.0.2.8"},
		{
			name: "forwarded header", remoteAddr: "192.0.2.8:4567", want: "192.0.2.8",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.9"},
		},
		{
			name: "real IP header", remoteAddr: "192.0.2.8:4567", want: "192.0.2.8",
			headers: map[string]string{"X-Real-IP": "203.0.113.9"},
		},
		{
			name: "forwarded chain", remoteAddr: "192.0.2.8:4567", want: "192.0.2.8",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.9, 198.51.100.1", "X-Real-IP": "203.0.113.10",
			},
		},
		{
			name: "IPv6", remoteAddr: "[2001:db8::8]:4567", want: "2001:db8::8",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.9"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), engine)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/email/code", nil)
			ctx.Request.RemoteAddr = test.remoteAddr
			for name, value := range test.headers {
				ctx.Request.Header.Set(name, value)
			}
			if got := ctx.ClientIP(); got != test.want {
				t.Fatalf("ClientIP() = %q, want direct peer %q", got, test.want)
			}
		})
	}
}
