package routers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/miebyte/authkit"
	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/middlewares"
	"github.com/superwhys/rent-manager/api/routers"
	"github.com/superwhys/rent-manager/config"
	apprbac "github.com/superwhys/rent-manager/internal/app/rbac"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
)

func newAdminHandler(service *apprbac.Service, accountID string) http.Handler {
	auth := authServiceStub{
		authenticate: func(context.Context, string) (*authkit.AuthenticatedSession, error) {
			return &authkit.AuthenticatedSession{Account: authkit.Account{ID: accountID}}, nil
		},
	}
	return ginutils.NewServerHandler(ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/admin"),
		ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(auth)),
		routers.AdminRouter(service),
	))
}

func TestAdminRouterListsBuiltinPermissionsWithoutRepository(t *testing.T) {
	handler := newAdminHandler(apprbac.NewService(nil, nil, nil, nil), strings.Repeat("a", 64))
	req := httptest.NewRequest(http.MethodGet, "/admin/permissions", nil)
	req.Header.Set("Authorization", "Bearer session")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var body struct {
		Code int                 `json:"code"`
		Data []map[string]string `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	want := domainrbac.BuiltinPermissions()
	if recorder.Code != http.StatusOK || body.Code != 0 || len(body.Data) != 10 {
		t.Fatalf("unexpected catalog: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	for i, permission := range body.Data {
		if len(permission) != 2 || permission["code"] != want[i].Code || permission["description"] != want[i].Description {
			t.Fatalf("catalog entry %d = %+v", i, permission)
		}
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin/permissions", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous catalog status = %d, want 401", recorder.Code)
	}
}

func TestAdminRouterRemovesPermissionCreation(t *testing.T) {
	handler := newAdminHandler(apprbac.NewService(nil, nil, nil, nil), strings.Repeat("a", 64))
	for _, path := range []string{"/admin/permissions/batch", "/admin/permissions"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"permissions":[{"code":"custom.permission","description":"自定义权限"}]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer session")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("removed %s route status = %d, want 404; body=%s", path, recorder.Code, recorder.Body.String())
		}
	}
}

func TestAdminRouterBindsOnlyBuiltinPermissionCodes(t *testing.T) {
	adminID := strings.Repeat("a", 64)
	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		for _, tt := range []struct {
			name string
			body string
			ok   bool
		}{
			{name: "known codes", body: `{"permission_codes":["room.view","unit.view","room.view"]}`, ok: true},
			{name: "unknown code", body: `{"permission_codes":["room.view","custom.permission"]}`},
			{name: "padded code", body: `{"permission_codes":[" room.view "]}`},
			{name: "uppercase code", body: `{"permission_codes":["ROOM.VIEW"]}`},
			{name: "empty codes", body: `{"permission_codes":[]}`},
			{name: "old permission IDs", body: `{"permission_ids":[1]}`},
		} {
			t.Run(method+"/"+tt.name, func(t *testing.T) {
				repo := &adminPermissionGroupRepository{}
				service := apprbac.NewService(repo, nil, nil, &config.AdminConfig{SuperAdminAccountIDs: []string{adminID}})
				handler := newAdminHandler(service, adminID)
				req := httptest.NewRequest(method, "/admin/permission-groups/7/permissions", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer session")
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)
				var body struct {
					Code int `json:"code"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if (body.Code == 0) != tt.ok || (repo.calls > 0) != tt.ok {
					t.Fatalf("binding result differs: calls=%d status=%d body=%s", repo.calls, recorder.Code, recorder.Body.String())
				}
				if tt.ok && (recorder.Code != http.StatusOK || repo.calls != 1 || repo.groupID != 7 ||
					!reflect.DeepEqual(repo.codes, []string{domainrbac.PermissionRoomView, domainrbac.PermissionUnitView})) {
					t.Fatalf("unexpected binding: repository=%+v body=%s", repo, recorder.Body.String())
				}
			})
		}
	}
}

type adminPermissionGroupRepository struct {
	domainrbac.Repository
	calls   int
	groupID int
	codes   []string
}

func (r *adminPermissionGroupRepository) AddPermissionsToPermissionGroup(_ context.Context, groupID int, codes []string) error {
	r.calls++
	r.groupID, r.codes = groupID, codes
	return nil
}

func (r *adminPermissionGroupRepository) RemovePermissionsFromPermissionGroup(_ context.Context, groupID int, codes []string) error {
	r.calls++
	r.groupID, r.codes = groupID, codes
	return nil
}
