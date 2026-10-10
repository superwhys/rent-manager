package mysql

import (
	"context"
	"errors"
	"sync"
	"testing"

	"gorm.io/gorm"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
)

func TestRBACRepositoryGrantUnitRolesRollsBack(t *testing.T) {
	for _, restore := range []bool{false, true} {
		name := "new grants"
		if restore {
			name = "restored and new grants"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			db := newTestDB(t)
			owner, target, unit := seedRBACUnit(t, db)
			repo := NewRBACRepository(db)
			roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{
				{AccountID: owner, Name: "first", IsActive: true},
				{AccountID: owner, Name: "second", IsActive: true},
			})
			if err != nil {
				t.Fatal(err)
			}
			groups, err := repo.CreatePermissionGroups(ctx, []*domainrbac.PermissionGroup{
				{Code: "update", Name: "修改单元", IsActive: true},
			})
			if err != nil {
				t.Fatal(err)
			}
			permissions, err := repo.CreatePermissions(ctx, []*domainrbac.Permission{
				{Code: domainrbac.PermissionUnitUpdate, Description: "修改单元", IsActive: true},
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := repo.AddPermissionGroupsToRole(ctx, owner, roles[0].ID, []int{groups[0].ID}); err != nil {
				t.Fatal(err)
			}
			if err := repo.AddPermissionsToPermissionGroup(ctx, groups[0].ID, []int{permissions[0].ID}); err != nil {
				t.Fatal(err)
			}
			params := domainrbac.UnitAccountRolesParams{
				OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID,
				RoleIDs: []int{roles[0].ID},
			}
			if restore {
				if _, err := repo.GrantUnitRoles(ctx, params); err != nil {
					t.Fatal(err)
				}
				if err := repo.RevokeUnitRoles(ctx, params); err != nil {
					t.Fatal(err)
				}
			}
			writeErr := errors.New("injected second role grant failure")
			writes := 0
			failAt := 2
			if restore {
				failAt = 1
			}
			if err := db.Callback().Create().Before("gorm:create").Register("rbac_test:fail_second_grant", func(tx *gorm.DB) {
				if tx.Statement.Table == "unit_account_roles" {
					writes++
					if writes == failAt {
						tx.AddError(writeErr)
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.Callback().Create().Remove("rbac_test:fail_second_grant") })
			params.RoleIDs = []int{roles[0].ID, roles[1].ID}
			if _, err := repo.GrantUnitRoles(ctx, params); !errors.Is(err, writeErr) {
				t.Fatalf("GrantUnitRoles error = %v, want %v", err, writeErr)
			}
			allowed, err := repo.HasUnitPermission(ctx, target, unit.ID, domainrbac.PermissionUnitUpdate)
			if err != nil || allowed {
				t.Fatalf("failed grant left effective permission: allowed=%v, err=%v", allowed, err)
			}
			grants, err := repo.ListUnitGrants(ctx, domainrbac.ListUnitGrantsQuery{
				OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if restore {
				if len(grants) != 1 || grants[0].Status != domainrbac.UnitGrantStatusDisabled {
					t.Fatalf("failed grant changed previous grants: %+v", grants)
				}
			} else if len(grants) != 0 {
				t.Fatalf("failed grant created grants: %+v", grants)
			}
		})
	}
}

func TestRBACRepositoryGrantUnitRolesConcurrentRetries(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	if db.Dialector.Name() == "mysql" {
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetMaxOpenConns(8)
	}
	owner, target, unit := seedRBACUnit(t, db)
	repo := NewRBACRepository(db)
	roles, err := repo.CreateRoles(ctx, []*domainrbac.Role{
		{AccountID: owner, Name: "first", IsActive: true},
		{AccountID: owner, Name: "second", IsActive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	params := domainrbac.UnitAccountRolesParams{
		OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID,
		RoleIDs: []int{roles[0].ID, roles[1].ID},
	}
	start := make(chan struct{})
	errCh := make(chan error, 8)
	var workers sync.WaitGroup
	for range cap(errCh) {
		workers.Go(func() {
			<-start
			_, err := repo.GrantUnitRoles(ctx, params)
			errCh <- err
		})
	}
	close(start)
	workers.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent retry failed: %v", err)
		}
	}
	grants, err := repo.ListUnitGrants(ctx, domainrbac.ListUnitGrantsQuery{
		OwnerAccountID: owner, TargetAccountID: target, UnitID: unit.ID,
	})
	if err != nil || len(grants) != 2 {
		t.Fatalf("concurrent grant count=%d, err=%v", len(grants), err)
	}
}
