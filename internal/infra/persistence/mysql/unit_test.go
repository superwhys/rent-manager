package mysql

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"

	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestUnitRepositoryOwnershipAndMeterItems(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	other := testAccountID("other")
	water := createUnitTestMeterItem(t, q, owner, "water")
	power := createUnitTestMeterItem(t, q, owner, "power")
	foreign := createUnitTestMeterItem(t, q, other, "water")
	repo := NewUnitRepository(db)
	unit, err := repo.Create(ctx, &domainunit.Unit{
		AccountID: owner, Name: "A", Address: "address", City: "city", Province: "province",
		Config: domainunit.UnitConfig{MeterReadingDay: 15}, DefaultMeterItemIDs: []int{water.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if unit.AccessType != domainunit.AccessTypeOwned || unit.Config.MeterReadingDay != 15 ||
		!reflect.DeepEqual(unit.DefaultMeterItemIDs, []int{water.ID}) {
		t.Fatalf("unexpected unit: %+v", unit)
	}
	if _, err := repo.Get(ctx, other, unit.ID); !errors.Is(err, domainunit.ErrNotFound) {
		t.Fatalf("foreign Get error = %v", err)
	}
	if _, err := repo.Create(ctx, unit); !errors.Is(err, domainunit.ErrNameAlreadyExists) {
		t.Fatalf("duplicate name error = %v", err)
	}
	if _, err := repo.Update(ctx, unit); err != nil {
		t.Fatalf("unchanged update: %v", err)
	}

	unit.DefaultMeterItemIDs = []int{power.ID}
	unit.Config.MeterReadingDay = 31
	unit, err = repo.Update(ctx, unit)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unit.DefaultMeterItemIDs, []int{power.ID}) ||
		unit.Config.MeterReadingDay != 31 {
		t.Fatalf("update did not replace relations/config: %+v", unit)
	}
	unit.DefaultMeterItemIDs = []int{foreign.ID}
	if _, err := repo.Update(ctx, unit); !errors.Is(err, domainunit.ErrMeterItemNotFound) {
		t.Fatalf("foreign meter item error = %v", err)
	}
	unit.DefaultMeterItemIDs = nil
	unit.AccountID = other
	if _, err := repo.Update(ctx, unit); !errors.Is(err, domainunit.ErrNotFound) {
		t.Fatalf("foreign Update error = %v", err)
	}
	if err := repo.Delete(ctx, other, unit.ID); !errors.Is(err, domainunit.ErrNotFound) {
		t.Fatalf("foreign Delete error = %v", err)
	}
	unit.AccountID = owner
	unit, err = repo.Update(ctx, unit)
	if err != nil || len(unit.DefaultMeterItemIDs) != 0 {
		t.Fatalf("clear meter items = %+v, %v", unit, err)
	}
	if err := repo.Delete(ctx, owner, unit.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, unit.ID); !errors.Is(err, domainunit.ErrNotFound) {
		t.Fatalf("deleted GetByID error = %v", err)
	}
}

func TestUnitRepositoryMeterItemTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	water := createUnitTestMeterItem(t, q, owner, "water")
	power := createUnitTestMeterItem(t, q, owner, "power")
	repo := NewUnitRepository(db)
	unit, err := repo.Create(ctx, &domainunit.Unit{
		AccountID: owner, Name: "before", Address: "address", City: "city", Province: "province",
		Config: domainunit.DefaultUnitConfig(), DefaultMeterItemIDs: []int{water.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeErr := errors.New("relation write failed")
	if err := db.Callback().
		Create().
		Before("gorm:create").
		Register("unit_test:fail_relation", func(tx *gorm.DB) {
			if tx.Statement.Table == "unit_meter_items" {
				tx.AddError(writeErr)
			}
		}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove("unit_test:fail_relation") })

	unit.Name = "after"
	unit.DefaultMeterItemIDs = []int{power.ID}
	if _, err := repo.Update(ctx, unit); !errors.Is(err, writeErr) {
		t.Fatalf("Update error = %v", err)
	}
	stored, err := repo.Get(ctx, owner, unit.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "before" || !reflect.DeepEqual(stored.DefaultMeterItemIDs, []int{water.ID}) {
		t.Fatalf("failed update was not rolled back: %+v", stored)
	}
	if _, err := repo.Create(ctx, unit); !errors.Is(err, writeErr) {
		t.Fatalf("Create error = %v", err)
	}
	count, err := q.Unit.WithContext(ctx).Count()
	if err != nil || count != 1 {
		t.Fatalf("failed create was not rolled back: count=%d, err=%v", count, err)
	}
}

func TestUnitRepositoryListScopes(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	viewer := testAccountID("viewer")
	owned := createUnitTestUnit(t, q, viewer, "owned")
	active := true
	inactive := false
	now := time.Now()
	role := &models.Role{AccountID: owner, Name: "active", IsActive: &active}
	secondRole := &models.Role{AccountID: owner, Name: "second", IsActive: &active}
	inactiveRole := &models.Role{AccountID: owner, Name: "inactive", IsActive: &inactive}
	deletedRole := &models.Role{
		AccountID: owner,
		Name:      "deleted",
		IsActive:  &active,
		DeletedAt: &now,
	}
	if err := q.Role.WithContext(ctx).
		Create(role, secondRole, inactiveRole, deletedRole); err != nil {
		t.Fatal(err)
	}
	granted := createUnitTestUnit(t, q, owner, "granted")
	fixtures := []struct {
		name      string
		roleID    int
		status    string
		deletedAt *time.Time
	}{
		{"disabled grant", role.ID, "disabled", nil},
		{"deleted grant", role.ID, "active", &now},
		{"inactive role", inactiveRole.ID, "active", nil},
		{"deleted role", deletedRole.ID, "active", nil},
	}
	for _, fixture := range fixtures {
		unit := createUnitTestUnit(t, q, owner, fixture.name)
		if err := q.UnitAccountRole.WithContext(ctx).Create(&models.UnitAccountRole{
			UnitID: unit.ID, AccountID: viewer, RoleID: fixture.roleID, GrantedByAccountID: owner,
			Status: fixture.status, DeletedAt: fixture.deletedAt,
		}); err != nil {
			t.Fatal(err)
		}
	}
	for _, roleID := range []int{role.ID, secondRole.ID} {
		if err := q.UnitAccountRole.WithContext(ctx).Create(&models.UnitAccountRole{
			UnitID:             granted.ID,
			AccountID:          viewer,
			RoleID:             roleID,
			GrantedByAccountID: owner,
			Status:             "active",
		}); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewUnitRepository(db)
	for _, test := range []struct {
		scope string
		ids   []int
	}{
		{domainunit.ListScopeOwned, []int{owned.ID}},
		{domainunit.ListScopeGranted, []int{granted.ID}},
		{domainunit.ListScopeAll, []int{granted.ID, owned.ID}},
		{"", []int{granted.ID, owned.ID}},
	} {
		t.Run(test.scope, func(t *testing.T) {
			units, err := repo.List(ctx, domainunit.ListQuery{AccountID: viewer, Scope: test.scope})
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]int, 0, len(units))
			for _, unit := range units {
				ids = append(ids, unit.ID)
				if unit.ID == granted.ID && unit.AccessType != domainunit.AccessTypeGranted {
					t.Fatalf("granted access type = %q", unit.AccessType)
				}
			}
			if !reflect.DeepEqual(ids, test.ids) {
				t.Fatalf("ids = %v, want %v", ids, test.ids)
			}
		})
	}
	if _, err := repo.List(
		ctx,
		domainunit.ListQuery{AccountID: viewer, Scope: "invalid"},
	); !errors.Is(
		err,
		domainunit.ErrInvalidListScope,
	) {
		t.Fatalf("invalid scope error = %v", err)
	}
	if err := repo.Delete(ctx, owner, granted.ID); err != nil {
		t.Fatal(err)
	}
	units, err := repo.List(
		ctx,
		domainunit.ListQuery{AccountID: viewer, Scope: domainunit.ListScopeGranted},
	)
	if err != nil || len(units) != 0 {
		t.Fatalf("deleted granted units = %+v, %v", units, err)
	}
}

func createUnitTestMeterItem(
	t *testing.T,
	q *query.Query,
	accountID string,
	code string,
) *models.MeterItem {
	t.Helper()
	item := &models.MeterItem{
		AccountID: accountID,
		Code:      code,
		Unit:      "unit",
		UnitPrice: 3,
		CalcType:  "DIFFERENCE",
	}
	if err := q.MeterItem.WithContext(context.Background()).Create(item); err != nil {
		t.Fatal(err)
	}
	return item
}

func createUnitTestUnit(t *testing.T, q *query.Query, accountID string, name string) *models.Unit {
	t.Helper()
	unit := &models.Unit{
		AccountID:  accountID,
		Name:       name,
		Address:    "address",
		City:       "city",
		Province:   "province",
		JSONExtend: `{"meter_reading_day":1}`,
	}
	if err := q.Unit.WithContext(context.Background()).Create(unit); err != nil {
		t.Fatal(err)
	}
	return unit
}
