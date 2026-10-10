package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
)

func TestRBACRepositoryRoles(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{
		{AccountID: owner, Name: "管理员", IsDefaultOwner: true, IsActive: true},
		{AccountID: owner, Name: "停用角色", IsActive: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if roles[1].IsActive {
		t.Fatal("创建时显式 false 被默认值覆盖")
	}
	if roles[0].CreatedAt.IsZero() || roles[0].UpdatedAt.IsZero() {
		t.Fatal("角色时间未自动写入")
	}
	if _, err := repo.CreateRoles(
		ctx,
		[]*domainrbac.Role{{AccountID: owner, Name: roles[0].Name}},
	); !errors.Is(
		err,
		domainrbac.ErrRoleAlreadyExists,
	) {
		t.Fatalf("重复角色错误 = %v", err)
	}
	if _, err := repo.GetRole(
		ctx,
		other,
		roles[0].ID,
	); !errors.Is(
		err,
		domainrbac.ErrRoleNotFound,
	) {
		t.Fatalf("跨账号读取错误 = %v", err)
	}
	inactive := false
	listed, err := repo.ListRoles(
		ctx,
		domainrbac.ListRolesQuery{AccountID: owner, IsActive: &inactive},
	)
	if err != nil || len(listed) != 1 || listed[0].ID != roles[1].ID {
		t.Fatalf("停用角色筛选 = %+v, %v", listed, err)
	}
	listed, err = repo.ListRoles(
		ctx,
		domainrbac.ListRolesQuery{AccountID: owner, Offset: 1, Limit: 1},
	)
	if err != nil || len(listed) != 1 || listed[0].ID != roles[0].ID {
		t.Fatalf("角色排序分页 = %+v, %v", listed, err)
	}
	previousUpdate := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := db.Model(&models.Role{}).
		Where("id = ?", roles[0].ID).
		UpdateColumn("updated_at", previousUpdate).
		Error; err != nil {
		t.Fatal(err)
	}
	roles[0].IsActive = false
	roles[0].IsDefaultOwner = false
	updated, err := repo.UpdateRole(ctx, roles[0])
	if err != nil || updated.IsActive || updated.IsDefaultOwner {
		t.Fatalf("角色零值更新 = %+v, %v", updated, err)
	}
	if !updated.UpdatedAt.After(previousUpdate) {
		t.Fatal("角色更新未刷新 updated_at")
	}
	if err := repo.DeleteRole(
		ctx,
		other,
		roles[0].ID,
	); !errors.Is(
		err,
		domainrbac.ErrRoleNotFound,
	) {
		t.Fatalf("跨账号删除错误 = %v", err)
	}
	if err := repo.DeleteRole(ctx, owner, roles[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetRole(
		ctx,
		owner,
		roles[0].ID,
	); !errors.Is(
		err,
		domainrbac.ErrRoleNotFound,
	) {
		t.Fatalf("已删除角色读取错误 = %v", err)
	}
	if _, err := repo.UpdateRole(ctx, roles[0]); !errors.Is(err, domainrbac.ErrRoleNotFound) {
		t.Fatalf("已删除角色更新错误 = %v", err)
	}
}

func TestRBACRepositoryPermissionsAndBindings(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, other, _ := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(
		ctx,
		[]*domainrbac.Role{{AccountID: owner, Name: "管理员", IsActive: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	description := "权限组描述"
	groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{
		{Code: "first", Name: "第一组", SortOrder: 2, IsActive: true, Description: &description},
		{Code: "second", Name: "第二组", SortOrder: 1, IsActive: false},
		{Code: "third", Name: "第三组", SortOrder: 1, IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if groups[1].IsActive || groups[1].Description != nil || groups[0].Description == nil ||
		*groups[0].Description != description {
		t.Fatalf("权限组 nullable/false 值未保留: %+v", groups)
	}
	permissions, err := repo.CreatePermissions(ctx, []*domainrbac.Permission{
		{Code: "unit.read", Description: "读取单元", IsActive: true},
		{Code: "unit.write", Description: "修改单元", IsActive: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if permissions[1].IsActive {
		t.Fatal("权限创建时显式 false 被默认值覆盖")
	}
	if _, err := repo.CreatePermissionGroups(
		ctx,
		[]*domainrbac.PermissionGroup{{Code: "first", Name: "重复组"}},
	); !errors.Is(
		err,
		domainrbac.ErrPermissionGroupAlreadyExists,
	) {
		t.Fatalf("重复权限组错误 = %v", err)
	}
	if _, err := repo.CreatePermissions(
		ctx,
		[]*domainrbac.Permission{{Code: "unit.read", Description: "重复权限"}},
	); !errors.Is(
		err,
		domainrbac.ErrPermissionAlreadyExists,
	) {
		t.Fatalf("重复权限错误 = %v", err)
	}
	listedGroups, err := repo.ListPermissionGroups(ctx)
	if err != nil || len(listedGroups) != 3 || listedGroups[0].ID != groups[2].ID ||
		listedGroups[1].ID != groups[1].ID {
		t.Fatalf("权限组排序 = %+v, %v", listedGroups, err)
	}
	listedPermissions, err := repo.ListPermissions(ctx)
	if err != nil || len(listedPermissions) != 2 || listedPermissions[0].ID != permissions[1].ID {
		t.Fatalf("权限排序 = %+v, %v", listedPermissions, err)
	}
	if err := repo.AddPermissionGroupsToRole(
		ctx,
		other,
		roles[0].ID,
		[]int{groups[0].ID},
	); !errors.Is(
		err,
		domainrbac.ErrRoleNotFound,
	) {
		t.Fatalf("跨账号绑定错误 = %v", err)
	}
	if err := repo.AddPermissionGroupsToRole(
		ctx,
		owner,
		roles[0].ID,
		[]int{99999},
	); !errors.Is(
		err,
		domainrbac.ErrPermissionGroupNotFound,
	) {
		t.Fatalf("不存在的权限组错误 = %v", err)
	}
	if err := repo.AddPermissionsToPermissionGroup(
		ctx,
		groups[0].ID,
		[]int{99999},
	); !errors.Is(
		err,
		domainrbac.ErrPermissionNotFound,
	) {
		t.Fatalf("不存在的权限错误 = %v", err)
	}
	for range 2 {
		if err := repo.AddPermissionGroupsToRole(
			ctx,
			owner,
			roles[0].ID,
			[]int{groups[0].ID},
		); err != nil {
			t.Fatal(err)
		}
		if err := repo.AddPermissionsToPermissionGroup(
			ctx,
			groups[0].ID,
			[]int{permissions[0].ID},
		); err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	if err := db.Model(&models.RolePermissionGroup{}).
		Count(&count).
		Error; err != nil ||
		count != 1 {
		t.Fatalf("重复角色绑定数量 = %d, %v", count, err)
	}
	if err := db.Model(&models.PermissionGroupPermission{}).
		Count(&count).
		Error; err != nil ||
		count != 1 {
		t.Fatalf("重复权限绑定数量 = %d, %v", count, err)
	}
	if err := repo.RemovePermissionGroupsFromRole(
		ctx,
		owner,
		roles[0].ID,
		[]int{groups[0].ID},
	); err != nil {
		t.Fatal(err)
	}
	if err := repo.RemovePermissionsFromPermissionGroup(
		ctx,
		groups[0].ID,
		[]int{permissions[0].ID},
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.RolePermissionGroup{}).
		Count(&count).
		Error; err != nil ||
		count != 0 {
		t.Fatalf("移除角色绑定数量 = %d, %v", count, err)
	}
	if err := db.Model(&models.PermissionGroupPermission{}).
		Count(&count).
		Error; err != nil ||
		count != 0 {
		t.Fatalf("移除权限绑定数量 = %d, %v", count, err)
	}
}

func TestRBACRepositoryUnitGrants(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, target, unit := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{
		{AccountID: owner, Name: "读取角色", IsActive: true},
		{AccountID: target, Name: "其他账号角色", IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	params := domainrbac.UnitAccountRolesParams{
		OwnerAccountID:  owner,
		TargetAccountID: target,
		UnitID:          unit.ID,
		RoleIDs:         []int{roles[0].ID},
	}
	for _, test := range []struct {
		name   string
		params domainrbac.UnitAccountRolesParams
		want   error
	}{
		{"单元不属于账号", domainrbac.UnitAccountRolesParams{OwnerAccountID: target, TargetAccountID: owner, UnitID: unit.ID, RoleIDs: params.RoleIDs}, domainrbac.ErrUnitNotInAccount},
		{"角色不属于账号", domainrbac.UnitAccountRolesParams{OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID, RoleIDs: []int{roles[1].ID}}, domainrbac.ErrRoleNotInAccount},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := repo.GrantUnitRoles(ctx, test.params); !errors.Is(err, test.want) {
				t.Fatalf("授权错误 = %v, want %v", err, test.want)
			}
		})
	}
	grants, err := repo.GrantUnitRoles(ctx, params)
	if err != nil || len(grants) != 1 {
		t.Fatalf("授权 = %+v, %v", grants, err)
	}
	if grants[0].Role == nil || grants[0].Role.ID != roles[0].ID {
		t.Fatalf("授权未加载角色: %+v", grants[0])
	}
	if grants[0].AccountID != target || grants[0].GrantedByAccountID != owner {
		t.Fatalf("授权双方账号 ID 未完整保存: %+v", grants[0])
	}
	if err := repo.RevokeUnitRoles(ctx, params); err != nil {
		t.Fatal(err)
	}
	disabled, err := repo.ListUnitGrants(
		ctx,
		domainrbac.ListUnitGrantsQuery{
			OwnerAccountID:  owner,
			UnitID:          unit.ID,
			TargetAccountID: target,
			Status:          domainrbac.UnitGrantStatusDisabled,
		},
	)
	if err != nil || len(disabled) != 1 {
		t.Fatalf("撤销后列表 = %+v, %v", disabled, err)
	}
	regranted, err := repo.GrantUnitRoles(ctx, params)
	if err != nil || len(regranted) != 1 || regranted[0].ID != grants[0].ID ||
		regranted[0].Status != domainrbac.UnitGrantStatusActive {
		t.Fatalf("恢复授权 = %+v, %v", regranted, err)
	}
	if err := repo.DeleteRole(ctx, owner, roles[0].ID); err != nil {
		t.Fatal(err)
	}
	disabled, err = repo.ListUnitGrants(
		ctx,
		domainrbac.ListUnitGrantsQuery{
			OwnerAccountID: owner,
			UnitID:         unit.ID,
			Status:         domainrbac.UnitGrantStatusDisabled,
		},
	)
	if err != nil || len(disabled) != 1 || disabled[0].Role.DeletedAt == nil {
		t.Fatalf("删除角色后的授权 = %+v, %v", disabled, err)
	}
}

func TestRBACRepositoryHasUnitPermission(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	owner, target, unit := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(
		ctx,
		[]*domainrbac.Role{{AccountID: owner, Name: "读取角色", IsActive: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := repo.CreatePermissionGroups(
		ctx,
		[]*domainrbac.PermissionGroup{{Code: "unit", Name: "单元权限", IsActive: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	permissions, err := repo.CreatePermissions(
		ctx,
		[]*domainrbac.Permission{{Code: "unit.read", Description: "读取单元", IsActive: true}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionGroupsToRole(
		ctx,
		owner,
		roles[0].ID,
		[]int{groups[0].ID},
	); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddPermissionsToPermissionGroup(
		ctx,
		groups[0].ID,
		[]int{permissions[0].ID},
	); err != nil {
		t.Fatal(err)
	}
	params := domainrbac.UnitAccountRolesParams{
		OwnerAccountID:  owner,
		TargetAccountID: target,
		UnitID:          unit.ID,
		RoleIDs:         []int{roles[0].ID},
	}
	grants, err := repo.GrantUnitRoles(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	assertPermission := func(t *testing.T, accountID string, code string, want bool) {
		t.Helper()
		got, err := repo.HasUnitPermission(ctx, accountID, unit.ID, code)
		if err != nil || got != want {
			t.Fatalf("HasUnitPermission(%q, %q) = %v, %v; want %v", accountID, code, got, err, want)
		}
	}
	assertPermission(t, owner, "任意权限", true)
	assertPermission(t, target, "unit.read", true)
	assertPermission(t, target, "unit.write", false)
	if _, err := repo.HasUnitPermission(
		ctx,
		target,
		99999,
		"unit.read",
	); !errors.Is(
		err,
		domainrbac.ErrUnitNotInAccount,
	) {
		t.Fatalf("不存在的单元错误 = %v", err)
	}
	for _, test := range []struct {
		name   string
		model  interface{}
		id     int
		column string
		value  interface{}
		reset  interface{}
	}{
		{"角色停用", &models.Role{}, roles[0].ID, "is_active", false, true},
		{"角色删除", &models.Role{}, roles[0].ID, "deleted_at", time.Now(), nil},
		{"权限组停用", &models.PermissionGroup{}, groups[0].ID, "is_active", false, true},
		{"权限组删除", &models.PermissionGroup{}, groups[0].ID, "deleted_at", time.Now(), nil},
		{"权限停用", &models.Permission{}, permissions[0].ID, "is_active", false, true},
		{"权限删除", &models.Permission{}, permissions[0].ID, "deleted_at", time.Now(), nil},
		{"授权停用", &models.UnitAccountRole{}, grants[0].ID, "status", domainrbac.UnitGrantStatusDisabled, domainrbac.UnitGrantStatusActive},
		{"授权删除", &models.UnitAccountRole{}, grants[0].ID, "deleted_at", time.Now(), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := db.Model(test.model).
				Where("id = ?", test.id).
				Update(test.column, test.value).
				Error; err != nil {
				t.Fatal(err)
			}
			assertPermission(t, target, "unit.read", false)
			if err := db.Model(test.model).
				Where("id = ?", test.id).
				Update(test.column, test.reset).
				Error; err != nil {
				t.Fatal(err)
			}
			assertPermission(t, target, "unit.read", true)
		})
	}
}

func seedRBACUnit(t *testing.T, db *gorm.DB) (string, string, *models.Unit) {
	t.Helper()
	owner, target := testAccountID("owner"), testAccountID("target")
	unit := &models.Unit{AccountID: owner, Name: "单元", Address: "测试地址", City: "深圳", Province: "广东"}
	if err := db.Create(unit).Error; err != nil {
		t.Fatal(err)
	}
	return owner, target, unit
}
