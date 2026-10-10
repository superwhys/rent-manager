package mysql

import (
	"context"
	"reflect"
	"testing"
	"time"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestRBACRepositoryReadsBindings(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, target, unit := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{
		{AccountID: owner, Name: "bound", IsActive: true},
		{AccountID: owner, Name: "empty", IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{
		{Code: "bound", Name: "绑定组", IsActive: true},
		{Code: "inactive", Name: "停用组", IsActive: false},
		{Code: "deleted", Name: "已删除组", IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	permissions, err := repo.CreatePermissions(ctx, []*domainrbac.Permission{
		{Code: "unit.view", Description: "查看单元", IsActive: true},
		{Code: "unit.update", Description: "修改单元", IsActive: false},
		{Code: "deleted", Description: "已删除权限", IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionGroupsToRole(ctx, owner, roles[0].ID,
		[]int{groups[2].ID, groups[1].ID, groups[0].ID}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionsToPermissionGroup(ctx, groups[0].ID,
		[]int{permissions[2].ID, permissions[1].ID, permissions[0].ID}); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.PermissionGroup{}).Where("id = ?", groups[2].ID).
		Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.Permission{}).Where("id = ?", permissions[2].ID).
		Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	assertRole := func(t *testing.T, role *domainrbac.Role, want []int) {
		t.Helper()
		if role == nil || !reflect.DeepEqual(role.PermissionGroupIDs, want) {
			t.Fatalf("role bindings = %+v, want %v", role, want)
		}
	}
	wantGroups := []int{groups[0].ID, groups[1].ID}
	role, err := repo.GetRole(ctx, owner, roles[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	assertRole(t, role, wantGroups)
	listed, err := repo.ListRoles(ctx, domainrbac.ListRolesQuery{AccountID: owner})
	if err != nil || len(listed) != 2 {
		t.Fatalf("ListRoles = %+v, %v", listed, err)
	}
	assertRole(t, listed[0], []int{})
	assertRole(t, listed[1], wantGroups)
	role.Name = "renamed"
	role, err = repo.UpdateRole(ctx, role)
	if err != nil {
		t.Fatal(err)
	}
	assertRole(t, role, wantGroups)
	grants, err := repo.GrantUnitRoles(ctx, domainrbac.UnitAccountRolesParams{
		OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID,
		RoleIDs: []int{roles[0].ID},
	})
	if err != nil || len(grants) != 1 {
		t.Fatalf("GrantUnitRoles = %+v, %v", grants, err)
	}
	assertRole(t, grants[0].Role, wantGroups)
	listedGrants, err := repo.ListUnitGrants(ctx, domainrbac.ListUnitGrantsQuery{
		OwnerAccountID: owner, UnitID: unit.ID,
	})
	if err != nil || len(listedGrants) != 1 {
		t.Fatalf("ListUnitGrants = %+v, %v", listedGrants, err)
	}
	assertRole(t, listedGrants[0].Role, wantGroups)
	listedGroups, err := repo.ListPermissionGroups(ctx)
	if err != nil || len(listedGroups) != 2 {
		t.Fatalf("ListPermissionGroups = %+v, %v", listedGroups, err)
	}
	if !reflect.DeepEqual(listedGroups[0].PermissionIDs, []int{}) ||
		!reflect.DeepEqual(listedGroups[1].PermissionIDs, []int{permissions[0].ID, permissions[1].ID}) {
		t.Fatalf("permission bindings = %+v, %+v", listedGroups[0], listedGroups[1])
	}
	if err := repo.RemovePermissionGroupsFromRole(ctx, owner, role.ID, wantGroups); err != nil {
		t.Fatal(err)
	}
	if err := repo.RemovePermissionsFromPermissionGroup(ctx, groups[0].ID,
		[]int{permissions[0].ID, permissions[1].ID}); err != nil {
		t.Fatal(err)
	}
	role, err = repo.GetRole(ctx, owner, role.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertRole(t, role, []int{})
	listedGroups, err = repo.ListPermissionGroups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range listedGroups {
		if !reflect.DeepEqual(group.PermissionIDs, []int{}) {
			t.Fatalf("removed permission bindings = %+v", group)
		}
	}
}
