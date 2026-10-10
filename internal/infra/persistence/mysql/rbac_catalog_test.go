package mysql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"gorm.io/gorm"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestFreshSchemaUsesBuiltinPermissionCodes(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	if db.Migrator().HasTable("permissions") {
		t.Fatal("固定权限目录不应创建 permissions 表")
	}
	if db.Migrator().HasColumn(&models.PermissionGroupPermission{}, "permission_id") ||
		!db.Migrator().HasColumn(&models.PermissionGroupPermission{}, "permission_code") {
		t.Fatal("权限组绑定应直接保存 permission_code")
	}
	owner, target, unit := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{{AccountID: owner, Name: "查看", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{{Code: "read", Name: "查看", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionsToPermissionGroup(ctx, groups[0].ID, []string{domainrbac.PermissionUnitView}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionGroupsToRole(ctx, owner, roles[0].ID, []int{groups[0].ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GrantUnitRoles(ctx, domainrbac.UnitAccountRolesParams{
		OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID, RoleIDs: []int{roles[0].ID},
	}); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		code string
		want bool
	}{
		{domainrbac.PermissionUnitView, true},
		{domainrbac.PermissionUnitUpdate, false},
		{"UNIT.VIEW", false},
		{" unit.view ", false},
		{"custom.permission", false},
	} {
		got, err := repo.HasUnitPermission(ctx, target, unit.ID, check.code)
		if err != nil || got != check.want {
			t.Fatalf("权限 %q = %v, %v; want %v", check.code, got, err, check.want)
		}
	}
	if err := db.Create(&models.PermissionGroupPermission{PermissionGroupID: groups[0].ID, PermissionCode: "custom.permission"}).Error; err != nil {
		t.Fatal(err)
	}
	if allowed, err := repo.HasUnitPermission(ctx, target, unit.ID, "custom.permission"); err != nil || allowed {
		t.Fatalf("数据库中的未知权限不得生效: allowed=%v, err=%v", allowed, err)
	}
	if err := db.Create(&models.PermissionGroupPermission{PermissionGroupID: groups[0].ID, PermissionCode: "ROOM.MANAGE"}).Error; err != nil {
		t.Fatal(err)
	}
	if allowed, err := repo.HasUnitPermission(ctx, target, unit.ID, domainrbac.PermissionRoomManage); err != nil || allowed {
		t.Fatalf("绑定编码大小写不同不得生效: allowed=%v, err=%v", allowed, err)
	}
}

func TestRBACRepositoryRejectsInvalidPermissionBindings(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	repo := NewRBACRepository(db)
	groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{{Code: "read", Name: "查看", IsActive: true}})
	if err != nil {
		t.Fatal(err)
	}
	groupID := groups[0].ID
	if err := repo.AddPermissionsToPermissionGroup(ctx, groupID, []string{domainrbac.PermissionUnitView}); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []struct {
		name string
		call func(context.Context, int, []string) error
	}{
		{"添加", repo.AddPermissionsToPermissionGroup},
		{"移除", repo.RemovePermissionsFromPermissionGroup},
	} {
		for _, codes := range [][]string{
			nil,
			{},
			{""},
			{"unknown"},
			{"UNIT.VIEW"},
			{domainrbac.PermissionUnitView, "unknown"},
		} {
			t.Run(operation.name+fmt.Sprint(codes), func(t *testing.T) {
				if err := operation.call(ctx, groupID, codes); !errors.Is(err, domainrbac.ErrPermissionNotFound) {
					t.Fatalf("无效权限 %v 错误 = %v", codes, err)
				}
				listed, err := repo.ListPermissionGroups(ctx)
				if err != nil || !reflect.DeepEqual(listed[0].PermissionCodes, []string{domainrbac.PermissionUnitView}) {
					t.Fatalf("无效操作修改了绑定: %+v, %v", listed, err)
				}
			})
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if err := operation.call(cancelled, groupID, []string{domainrbac.PermissionUnitView}); !errors.Is(err, context.Canceled) {
			t.Fatalf("%s未传递取消context: %v", operation.name, err)
		}
	}
}

func TestRBACRepositoryPermissionBindingTransactionRollsBack(t *testing.T) {
	for _, remove := range []bool{false, true} {
		name := "添加"
		if remove {
			name = "移除"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			db := newTestDB(t)
			repo := NewRBACRepository(db)
			groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{{Code: "read", Name: "查看", IsActive: true}})
			if err != nil {
				t.Fatal(err)
			}
			groupID := groups[0].ID
			if remove {
				if err := repo.AddPermissionsToPermissionGroup(ctx, groupID, []string{domainrbac.PermissionUnitView}); err != nil {
					t.Fatal(err)
				}
			}
			writeErr := errors.New("injected permission group update failure")
			if err := db.Callback().Update().Before("gorm:update").Register("rbac_test:fail_group_update", func(tx *gorm.DB) {
				if tx.Statement.Table == "permission_groups" {
					tx.AddError(writeErr)
				}
			}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.Callback().Update().Remove("rbac_test:fail_group_update") })
			operation := repo.AddPermissionsToPermissionGroup
			if remove {
				operation = repo.RemovePermissionsFromPermissionGroup
			}
			if err := operation(ctx, groupID, []string{domainrbac.PermissionUnitView}); !errors.Is(err, writeErr) {
				t.Fatalf("事务错误 = %v, want %v", err, writeErr)
			}
			listed, err := repo.ListPermissionGroups(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{}
			if remove {
				want = []string{domainrbac.PermissionUnitView}
			}
			if !reflect.DeepEqual(listed[0].PermissionCodes, want) {
				t.Fatalf("失败事务修改了绑定: %+v; want %v", listed, want)
			}
		})
	}
}
