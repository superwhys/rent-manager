package routers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miebyte/goutils/ginutils"

	"github.com/superwhys/rent-manager/api/routers"
	appcore "github.com/superwhys/rent-manager/internal/app"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestDictionaryRoutesReadUnitOwnerDefinitions(t *testing.T) {
	fixture := newDictionaryFixture()
	for _, path := range []string{"/room-types?unit_id=7", "/meter-items?unit_id=7"} {
		t.Run(path, func(t *testing.T) {
			body := dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, path, "")
			var items []struct {
				ID        int    `json:"id"`
				AccountID string `json:"account_id"`
			}
			if err := json.Unmarshal(body.Data, &items); err != nil {
				t.Fatal(err)
			}
			if body.Code != 0 || len(items) != 1 || items[0].AccountID != fixture.owner {
				t.Fatalf("楼栋字典未返回主账号定义：%+v", body)
			}
		})
	}
}

func TestDictionaryRoutesEnforceReadScope(t *testing.T) {
	for _, test := range []struct {
		path        string
		wantAccount string
		wantID      int
		wantCode    int
	}{
		{path: "/room-types/1?unit_id=7", wantAccount: "owner", wantID: 1},
		{path: "/meter-items/1?unit_id=7", wantAccount: "owner", wantID: 1},
		{path: "/room-types/2", wantAccount: "collaborator", wantID: 2},
		{path: "/meter-items/2", wantAccount: "collaborator", wantID: 2},
		{path: "/room-types/1?unit_id=7&RoomTypeID=2", wantAccount: "owner", wantID: 1},
		{path: "/meter-items/1?unit_id=7&MeterItemID=2", wantAccount: "owner", wantID: 1},
		{path: "/room-types/1", wantCode: domainroomtype.ErrNotFound.Code()},
		{path: "/meter-items/1", wantCode: domainmeteritem.ErrNotFound.Code()},
		{path: "/room-types/2?unit_id=7", wantCode: domainroomtype.ErrNotFound.Code()},
		{path: "/meter-items/2?unit_id=7", wantCode: domainmeteritem.ErrNotFound.Code()},
		{path: "/room-types?unit_id=8", wantCode: domainrbac.ErrForbidden.Code()},
		{path: "/meter-items?unit_id=8", wantCode: domainrbac.ErrForbidden.Code()},
	} {
		t.Run(test.path, func(t *testing.T) {
			fixture := newDictionaryFixture()
			body := dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, test.path, "")
			if body.Code != test.wantCode {
				t.Fatalf("响应错误码 = %d, want %d", body.Code, test.wantCode)
			}
			if test.wantCode != 0 {
				return
			}
			var item struct {
				ID        int    `json:"id"`
				AccountID string `json:"account_id"`
			}
			if err := json.Unmarshal(body.Data, &item); err != nil {
				t.Fatal(err)
			}
			wantAccount := fixture.owner
			if test.wantAccount == "collaborator" {
				wantAccount = fixture.collaborator
			}
			if item.ID != test.wantID || item.AccountID != wantAccount {
				t.Fatalf("读取了错误的字典定义：%+v", item)
			}
		})
	}
}

func TestDictionaryRoutesRecheckPermissionAfterRevocation(t *testing.T) {
	for _, prefix := range []string{"/room-types", "/meter-items"} {
		t.Run(prefix, func(t *testing.T) {
			fixture := newDictionaryFixture()
			body := dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, prefix+"/1?unit_id=7", "")
			if body.Code != 0 {
				t.Fatalf("已授权读取失败：%+v", body)
			}
			fixture.permissions.allowed = nil
			reads := fixture.roomTypes.reads + fixture.meterItems.reads
			unitReads := fixture.units.calls
			for _, path := range []string{prefix + "?unit_id=7", prefix + "/1?unit_id=7"} {
				body = dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, path, "")
				if body.Code != domainrbac.ErrForbidden.Code() {
					t.Fatalf("撤权后仍允许读取：%+v", body)
				}
			}
			if fixture.roomTypes.reads+fixture.meterItems.reads != reads || fixture.units.calls != unitReads {
				t.Fatal("权限拒绝后仍然查询楼栋或共享字典")
			}
		})
	}
}

func TestDictionaryRoutesRequireSpecificViewPermission(t *testing.T) {
	fixture := newDictionaryFixture()
	fixture.permissions.allowed = map[string]bool{
		domainrbac.PermissionRoomTypeView:    true,
		domainrbac.PermissionRoomTypeManage:  true,
		domainrbac.PermissionMeterItemManage: true,
	}
	for _, path := range []string{"/meter-items?unit_id=7", "/meter-items/1?unit_id=7"} {
		body := dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, path, "")
		if body.Code != domainrbac.ErrForbidden.Code() {
			t.Fatalf("不相关权限意外授予收费项读取：%+v", body)
		}
	}
}

func TestDictionaryRoutesRejectInvalidUnitQuery(t *testing.T) {
	fixture := newDictionaryFixture()
	for _, prefix := range []string{"/room-types", "/room-types/1", "/meter-items", "/meter-items/1"} {
		for _, value := range []string{"-1", "abc"} {
			path := prefix + "?unit_id=" + value
			body := dictionaryRequest(t, fixture.handler, fixture.collaborator, http.MethodGet, path, "")
			if body.Code == 0 {
				t.Fatalf("非法楼栋参数未拒绝：%s", path)
			}
		}
	}
	if fixture.permissions.calls != 0 || fixture.roomTypes.reads+fixture.meterItems.reads != 0 {
		t.Fatal("非法请求进入了服务查询")
	}
}

func TestDictionaryRoutesDoNotGrantSharedDefinitionWrites(t *testing.T) {
	fixture := newDictionaryFixture()
	fixture.permissions.allowed[domainrbac.PermissionRoomTypeManage] = true
	fixture.permissions.allowed[domainrbac.PermissionMeterItemManage] = true
	for _, test := range []struct {
		method   string
		path     string
		body     string
		wantCode int
	}{
		{http.MethodPut, "/room-types/1?unit_id=7", `{"name":"改名"}`, domainroomtype.ErrNotFound.Code()},
		{http.MethodDelete, "/room-types/1?unit_id=7", "", domainroomtype.ErrNotFound.Code()},
		{http.MethodPut, "/meter-items/1?unit_id=7", `{"code":"水费","unit":"吨","unit_price":9}`, domainmeteritem.ErrNotFound.Code()},
	} {
		body := dictionaryRequest(t, fixture.handler, fixture.collaborator, test.method, test.path, test.body)
		if body.Code != test.wantCode {
			t.Fatalf("共享定义写入未保持所有者边界：%+v", body)
		}
	}
}

type dictionaryFixture struct {
	handler      http.Handler
	owner        string
	collaborator string
	permissions  *dictionaryRBACRepo
	units        *dictionaryUnitRepo
	roomTypes    *dictionaryRoomTypeRepo
	meterItems   *dictionaryMeterItemRepo
}

func newDictionaryFixture() dictionaryFixture {
	owner, collaborator := strings.Repeat("a", 64), strings.Repeat("b", 64)
	roomTypes := &dictionaryRoomTypeRepo{records: []*domainroomtype.RoomType{
		{ID: 1, AccountID: owner, Name: "主账号房型"},
		{ID: 2, AccountID: collaborator, Name: "个人房型"},
	}}
	meterItems := &dictionaryMeterItemRepo{records: []*domainmeteritem.MeterItem{
		{ID: 1, AccountID: owner, Code: "水费"},
		{ID: 2, AccountID: collaborator, Code: "个人水费"},
	}}
	units := &dictionaryUnitRepo{owner: owner}
	permissions := &dictionaryRBACRepo{accountID: collaborator, allowed: map[string]bool{
		domainrbac.PermissionRoomTypeView:  true,
		domainrbac.PermissionMeterItemView: true,
	}}
	handler := ginutils.NewServerHandler(
		ginutils.WithGroupHandlers(ginutils.WithPrefix("/room-types"),
			routers.RoomTypeRouter(approomtype.NewService(roomTypes, units, permissions))),
		ginutils.WithGroupHandlers(ginutils.WithPrefix("/meter-items"),
			routers.MeterItemRouter(appmeteritem.NewService(meterItems, units, permissions))),
	)
	return dictionaryFixture{handler: handler, owner: owner, collaborator: collaborator,
		permissions: permissions, units: units, roomTypes: roomTypes, meterItems: meterItems}
}

type dictionaryHTTPResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func dictionaryRequest(t *testing.T, handler http.Handler, accountID, method, path, requestBody string) dictionaryHTTPResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appcore.SetAccountIDToContext(req.Context(), accountID))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var body dictionaryHTTPResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("无法解析 HTTP %d 响应 %s：%v", recorder.Code, recorder.Body.String(), err)
	}
	return body
}

type dictionaryRoomTypeRepo struct {
	domainroomtype.Repository
	records []*domainroomtype.RoomType
	reads   int
}

func (r *dictionaryRoomTypeRepo) List(_ context.Context, query domainroomtype.ListQuery) ([]*domainroomtype.RoomType, error) {
	r.reads++
	items := make([]*domainroomtype.RoomType, 0)
	for _, item := range r.records {
		if item.AccountID == query.AccountID {
			items = append(items, item)
		}
	}
	return items, nil
}

type dictionaryMeterItemRepo struct {
	domainmeteritem.Repository
	records []*domainmeteritem.MeterItem
	reads   int
}

func (r *dictionaryMeterItemRepo) List(_ context.Context, query domainmeteritem.ListQuery) ([]*domainmeteritem.MeterItem, error) {
	r.reads++
	items := make([]*domainmeteritem.MeterItem, 0)
	for _, item := range r.records {
		if item.AccountID == query.AccountID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *dictionaryRoomTypeRepo) Get(_ context.Context, accountID string, id int) (*domainroomtype.RoomType, error) {
	r.reads++
	for _, item := range r.records {
		if item.AccountID == accountID && item.ID == id {
			return item, nil
		}
	}
	return nil, domainroomtype.ErrNotFound
}

func (r *dictionaryRoomTypeRepo) Update(ctx context.Context, item *domainroomtype.RoomType) (*domainroomtype.RoomType, error) {
	return r.Get(ctx, item.AccountID, item.ID)
}

func (r *dictionaryRoomTypeRepo) Delete(ctx context.Context, accountID string, id int) error {
	_, err := r.Get(ctx, accountID, id)
	return err
}

func (r *dictionaryMeterItemRepo) Get(_ context.Context, accountID string, id int) (*domainmeteritem.MeterItem, error) {
	r.reads++
	for _, item := range r.records {
		if item.AccountID == accountID && item.ID == id {
			return item, nil
		}
	}
	return nil, domainmeteritem.ErrNotFound
}

func (r *dictionaryMeterItemRepo) Update(ctx context.Context, item *domainmeteritem.MeterItem) (*domainmeteritem.MeterItem, error) {
	return r.Get(ctx, item.AccountID, item.ID)
}

type dictionaryUnitRepo struct {
	domainunit.Repository
	owner string
	calls int
}

func (r *dictionaryUnitRepo) GetByID(_ context.Context, id int) (*domainunit.Unit, error) {
	r.calls++
	if id != 7 {
		return nil, domainunit.ErrNotFound
	}
	return &domainunit.Unit{ID: id, AccountID: r.owner}, nil
}

type dictionaryRBACRepo struct {
	domainrbac.Repository
	accountID string
	allowed   map[string]bool
	calls     int
}

func (r *dictionaryRBACRepo) HasUnitPermission(_ context.Context, accountID string, unitID int, code string) (bool, error) {
	r.calls++
	return accountID == r.accountID && unitID == 7 && r.allowed[code], nil
}
