package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	apphome "github.com/superwhys/rent-manager/internal/app/home"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	approomtype "github.com/superwhys/rent-manager/internal/app/roomtype"
	appunit "github.com/superwhys/rent-manager/internal/app/unit"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestUnitRoleRevocationAppliesAcrossReadServices(t *testing.T) {
	db := newTestDB(t)
	owner, target, unit := seedRBACUnit(t, db)
	ctx := appcore.SetAccountIDToContext(context.Background(), target)
	typeRecord := &models.RoomType{AccountID: owner, Name: "一居室"}
	item := &models.MeterItem{AccountID: owner, Code: "water", Unit: "吨", UnitPrice: 100}
	for _, record := range []any{typeRecord, item} {
		if err := db.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	room := &models.Room{UnitID: unit.ID, RoomTypeID: typeRecord.ID, RoomNo: "101", Price: 123456, Size: 20, IsActive: new(false)}
	if err := db.Create(room).Error; err != nil {
		t.Fatal(err)
	}
	bill := &models.RentBill{RoomID: room.ID, PeriodMonth: time.Now().Format("2006-01"), TotalAmount: 123456, Status: domainrentbill.StatusPaid}
	if err := db.Create(bill).Error; err != nil {
		t.Fatal(err)
	}

	repos := NewRepositoryFactory(db).Repositories()
	roles, err := repos.RBAC.CreateRoles(ctx, []*domainrbac.Role{{AccountID: owner, Name: "协作者", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := repos.RBAC.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{{Code: "read", Name: "查看", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	permissionCodes := []string{domainrbac.PermissionUnitView, domainrbac.PermissionRoomView, domainrbac.PermissionRentBillView, domainrbac.PermissionRoomTypeView, domainrbac.PermissionMeterItemView}
	if err := repos.RBAC.AddPermissionsToPermissionGroup(ctx, groups[0].ID, permissionCodes); err != nil {
		t.Fatal(err)
	}
	if err := repos.RBAC.AddPermissionGroupsToRole(ctx, owner, roles[0].ID, []int{groups[0].ID}); err != nil {
		t.Fatal(err)
	}
	grant := domainrbac.UnitAccountRolesParams{OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID, RoleIDs: []int{roles[0].ID}}
	if _, err := repos.RBAC.GrantUnitRoles(ctx, grant); err != nil {
		t.Fatal(err)
	}

	home := apphome.NewService(repos.Unit, repos.Room, repos.RentBill, repos.MeterReading, repos.RBAC)
	units := appunit.NewService(repos.Unit, repos.Room, repos.RBAC)
	types := approomtype.NewService(repos.RoomType, repos.Unit, repos.RBAC)
	items := appmeteritem.NewService(repos.MeterItem, repos.Unit, repos.RBAC)
	summary, err := home.GetHomeSummary(ctx)
	if err != nil || summary.CurrentMonthIncomeCents != 123456 || summary.RoomCount != 1 {
		t.Fatalf("authorized summary = %+v, %v", summary, err)
	}
	listed, err := units.ListUnits(ctx, &appunit.ListUnitsQuery{Scope: "granted"})
	if err != nil || len(listed) != 1 {
		t.Fatalf("authorized units = %+v, %v", listed, err)
	}
	if response := appunit.ToUnitSummaryResponse(listed[0]); response.Address != unit.Address || response.RoomCount == nil || *response.RoomCount != 1 {
		t.Fatalf("authorized details = %+v", response)
	}
	if options, err := types.ListRoomTypes(ctx, &approomtype.ListRoomTypesQuery{UnitID: unit.ID}); err != nil || len(options) != 1 || options[0].ID != typeRecord.ID {
		t.Fatalf("authorized room types = %+v, %v", options, err)
	}
	if options, err := items.ListMeterItems(ctx, &appmeteritem.ListMeterItemsQuery{UnitID: unit.ID}); err != nil || len(options) != 1 || options[0].ID != item.ID {
		t.Fatalf("authorized meter items = %+v, %v", options, err)
	}

	// Removing bindings must apply on the next read, while the active grant still provides a minimal entry.
	if err := repos.RBAC.RemovePermissionGroupsFromRole(ctx, owner, roles[0].ID, []int{groups[0].ID}); err != nil {
		t.Fatal(err)
	}
	summary, err = home.GetHomeSummary(ctx)
	if err != nil || summary.CurrentMonthIncomeCents != 0 || summary.RoomCount != 0 || len(summary.Todos) != 0 {
		t.Fatalf("revoked summary = %+v, %v", summary, err)
	}
	if _, err := units.GetUnitInfo(ctx, unit.ID); !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("revoked unit detail = %v", err)
	}
	listed, err = units.ListUnits(ctx, &appunit.ListUnitsQuery{Scope: "granted"})
	if err != nil || len(listed) != 1 {
		t.Fatalf("minimal units = %+v, %v", listed, err)
	}
	data, err := json.Marshal(appunit.ToUnitSummaryResponse(listed[0]))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 4 || fields["id"] != float64(unit.ID) || fields["name"] != unit.Name || fields["access_type"] != "granted" || fields["is_owner"] != false {
		t.Fatalf("unexpected minimal entry: %s", data)
	}
	if _, err := types.ListRoomTypes(ctx, &approomtype.ListRoomTypesQuery{UnitID: unit.ID}); !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("revoked room types = %v", err)
	}
	if _, err := items.ListMeterItems(ctx, &appmeteritem.ListMeterItemsQuery{UnitID: unit.ID}); !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("revoked meter items = %v", err)
	}
	if err := repos.RBAC.RevokeUnitRoles(ctx, grant); err != nil {
		t.Fatal(err)
	}
	if listed, err := units.ListUnits(ctx, &appunit.ListUnitsQuery{Scope: "granted"}); err != nil || len(listed) != 0 {
		t.Fatalf("revoked grants still visible = %+v, %v", listed, err)
	}
	ownerCtx := appcore.SetAccountIDToContext(ctx, owner)
	if summary, err := home.GetHomeSummary(ownerCtx); err != nil || summary.CurrentMonthIncomeCents != 123456 {
		t.Fatalf("owner access changed = %+v, %v", summary, err)
	}
}
