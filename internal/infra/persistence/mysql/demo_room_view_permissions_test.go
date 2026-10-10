package mysql

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestDemoRoomViewPermissionRepair(t *testing.T) {
	repairSQL, err := os.ReadFile("../../../../testdata/fix_demo_room_view_permissions.sql")
	if err != nil {
		t.Fatal(err)
	}
	db := newTestDB(t)
	owner, reader, unit := seedRBACUnit(t, db)
	finance, other, revoked := testAccountID("finance"), testAccountID("other"), testAccountID("revoked")
	roomType := &models.RoomType{AccountID: owner, Name: "一居室"}
	meterItem := &models.MeterItem{AccountID: owner, Code: "水费", Unit: "吨"}
	for _, record := range []any{roomType, meterItem} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	scenarios := []struct {
		groupCode   string
		accountID   string
		codes       []string
		beforeItems bool
		afterRead   bool
	}{
		{"rent_demo.readings", reader, []string{
			domainrbac.PermissionUnitView, domainrbac.PermissionRoomView, domainrbac.PermissionMeterItemView,
			domainrbac.PermissionMeterReadingView, domainrbac.PermissionMeterReadingManage,
		}, true, true},
		{"rent_demo.finance", finance, []string{
			domainrbac.PermissionUnitView, domainrbac.PermissionRoomView,
			domainrbac.PermissionRentBillView, domainrbac.PermissionRentBillManage,
		}, false, true},
		{"other", other, []string{domainrbac.PermissionRoomView}, false, false},
	}
	for _, scenario := range scenarios {
		group := &models.PermissionGroup{Code: scenario.groupCode, Name: scenario.groupCode}
		role := &models.Role{AccountID: owner, Name: scenario.groupCode}
		for _, record := range []any{group, role} {
			if err := db.Create(record).Error; err != nil {
				t.Fatal(err)
			}
		}
		bindings := []any{
			&models.RolePermissionGroup{RoleID: role.ID, PermissionGroupID: group.ID},
			&models.UnitAccountRole{UnitID: unit.ID, AccountID: scenario.accountID, RoleID: role.ID, GrantedByAccountID: owner},
		}
		for _, code := range scenario.codes {
			bindings = append(bindings, &models.PermissionGroupPermission{PermissionGroupID: group.ID, PermissionCode: code})
		}
		if scenario.accountID == finance {
			bindings = append(bindings, &models.UnitAccountRole{
				UnitID: unit.ID, AccountID: revoked, RoleID: role.ID, GrantedByAccountID: owner,
				Status: domainrbac.UnitGrantStatusDisabled,
			})
		}
		for _, record := range bindings {
			if err := db.Create(record).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	repos := NewRepositoryFactory(db).Repositories()
	roomTypes := approomtype.NewService(repos.RoomType, repos.Unit, repos.RBAC)
	meterItems := appmeteritem.NewService(repos.MeterItem, repos.Unit, repos.RBAC)
	assertDirectories := func(accountID string, wantTypes, wantItems bool) {
		t.Helper()
		ctx := appcore.SetAccountIDToContext(context.Background(), accountID)
		types, err := roomTypes.ListRoomTypes(ctx, &approomtype.ListRoomTypesQuery{UnitID: unit.ID})
		if wantTypes {
			if err != nil || len(types) != 1 || types[0].ID != roomType.ID {
				t.Fatalf("共享房型目录 = %+v, %v", types, err)
			}
		} else if !errors.Is(err, domainrbac.ErrForbidden) {
			t.Fatalf("未授权房型目录错误 = %v", err)
		}
		items, err := meterItems.ListMeterItems(ctx, &appmeteritem.ListMeterItemsQuery{UnitID: unit.ID})
		if wantItems {
			if err != nil || len(items) != 1 || items[0].ID != meterItem.ID {
				t.Fatalf("共享收费项目录 = %+v, %v", items, err)
			}
		} else if !errors.Is(err, domainrbac.ErrForbidden) {
			t.Fatalf("未授权收费项目录错误 = %v", err)
		}
	}
	for _, scenario := range scenarios {
		assertDirectories(scenario.accountID, false, scenario.beforeItems)
	}
	for _, wantAdded := range []int64{3, 0} {
		result := db.Exec(string(repairSQL))
		if result.Error != nil || result.RowsAffected != wantAdded {
			t.Fatalf("修复新增绑定 = %d, %v; want %d", result.RowsAffected, result.Error, wantAdded)
		}
	}
	var bindingCount int64
	if err := db.Model(&models.PermissionGroupPermission{}).Count(&bindingCount).Error; err != nil || bindingCount != 13 {
		t.Fatalf("修复后绑定数 = %d, %v; want 13", bindingCount, err)
	}
	for _, scenario := range scenarios {
		assertDirectories(scenario.accountID, scenario.afterRead, scenario.afterRead)
	}
	assertDirectories(revoked, false, false)
	for _, check := range []struct {
		code    string
		allowed bool
	}{
		{domainrbac.PermissionRoomView, true},
		{domainrbac.PermissionRentBillView, true},
		{domainrbac.PermissionRentBillManage, true},
		{domainrbac.PermissionRoomManage, false},
		{domainrbac.PermissionMeterReadingManage, false},
	} {
		allowed, err := repos.RBAC.HasUnitPermission(context.Background(), finance, unit.ID, check.code)
		if err != nil || allowed != check.allowed {
			t.Fatalf("财务权限 %s = %v, %v; want %v", check.code, allowed, err, check.allowed)
		}
	}

	deletedAt := time.Now()
	for _, scenario := range []struct {
		name  string
		group models.PermissionGroup
	}{
		{"停用权限组", models.PermissionGroup{IsActive: new(false)}},
		{"删除权限组", models.PermissionGroup{DeletedAt: &deletedAt}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			db := newTestDB(t)
			group := scenario.group
			group.Code, group.Name = "rent_demo.finance", "财务"
			if err := db.Create(&group).Error; err != nil {
				t.Fatal(err)
			}
			result := db.Exec(string(repairSQL))
			if result.Error != nil || result.RowsAffected != 0 {
				t.Fatalf("停用或删除记录新增绑定 = %d, %v; want 0", result.RowsAffected, result.Error)
			}
		})
	}

}
