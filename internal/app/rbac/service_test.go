package rbac

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/miebyte/authkit"

	"github.com/superwhys/rent-manager/config"
	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestListPermissionsUsesBuiltinCatalogWithoutRepository(t *testing.T) {
	service := NewService(nil, nil, nil, nil)
	permissions, err := service.ListPermissions(context.Background())
	if err != nil || !reflect.DeepEqual(permissions, domainrbac.BuiltinPermissions()) {
		t.Fatalf("ListPermissions = %+v, error = %v", permissions, err)
	}
	permissions[0].Code = "custom.permission"
	next, err := service.ListPermissions(context.Background())
	if err != nil || next[0].Code != domainrbac.PermissionUnitView {
		t.Fatalf("permission catalog was mutated: %+v, error = %v", next, err)
	}
}

func TestPermissionGroupBindingsRequireBuiltinCodesAndSuperAdmin(t *testing.T) {
	adminID := strings.Repeat("a", 64)
	storageErr := errors.New("权限组仓储不可用")
	for _, operation := range []struct {
		name string
		call func(*Service, context.Context, *PermissionGroupPermissionsCommand) error
	}{
		{name: "add", call: (*Service).AddPermissionsToPermissionGroup},
		{name: "remove", call: (*Service).RemovePermissionsFromPermissionGroup},
	} {
		for _, tt := range []struct {
			name      string
			accountID string
			codes     []string
			repoErr   error
			wantErr   error
			wantCall  bool
		}{
			{name: "deduplicates", accountID: adminID, codes: []string{domainrbac.PermissionRoomView, domainrbac.PermissionUnitView, domainrbac.PermissionRoomView}, wantCall: true},
			{name: "nil codes", accountID: adminID, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "empty codes", accountID: adminID, codes: []string{}, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "empty code", accountID: adminID, codes: []string{""}, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "unknown code", accountID: adminID, codes: []string{domainrbac.PermissionRoomView, "custom.permission"}, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "uppercase code", accountID: adminID, codes: []string{"ROOM.VIEW"}, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "padded code", accountID: adminID, codes: []string{" room.view "}, wantErr: domainrbac.ErrPermissionNotFound},
			{name: "ordinary account", accountID: strings.Repeat("b", 64), codes: []string{domainrbac.PermissionRoomView}, wantErr: domainrbac.ErrForbidden},
			{name: "repository error", accountID: adminID, codes: []string{domainrbac.PermissionRoomView}, repoErr: storageErr, wantErr: storageErr, wantCall: true},
		} {
			t.Run(operation.name+"/"+tt.name, func(t *testing.T) {
				repo := &permissionGroupRepository{err: tt.repoErr}
				service := NewService(repo, nil, nil, &config.AdminConfig{SuperAdminAccountIDs: []string{adminID}})
				ctx := appcore.SetAccountIDToContext(context.Background(), tt.accountID)
				err := operation.call(service, ctx, &PermissionGroupPermissionsCommand{PermissionGroupID: 7, PermissionCodes: tt.codes})
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("binding error = %v, want %v", err, tt.wantErr)
				}
				if (repo.calls > 0) != tt.wantCall {
					t.Fatalf("repository calls = %d, want called=%v", repo.calls, tt.wantCall)
				}
				if tt.wantCall {
					wantCodes := []string{domainrbac.PermissionRoomView}
					if tt.name == "deduplicates" {
						wantCodes = append(wantCodes, domainrbac.PermissionUnitView)
					}
					if repo.calls != 1 || repo.operation != operation.name || repo.groupID != 7 || !reflect.DeepEqual(repo.codes, wantCodes) {
						t.Fatalf("unexpected repository binding: %+v", repo)
					}
				}
			})
		}
	}
}

type permissionGroupRepository struct {
	domainrbac.Repository
	calls     int
	operation string
	groupID   int
	codes     []string
	err       error
}

func (r *permissionGroupRepository) AddPermissionsToPermissionGroup(_ context.Context, groupID int, codes []string) error {
	r.calls++
	r.operation, r.groupID, r.codes = "add", groupID, codes
	return r.err
}

func (r *permissionGroupRepository) RemovePermissionsFromPermissionGroup(_ context.Context, groupID int, codes []string) error {
	r.calls++
	r.operation, r.groupID, r.codes = "remove", groupID, codes
	return r.err
}

func TestGrantUnitRolesUsesAuthkitAccountAfterOwnershipCheck(t *testing.T) {
	ownerID, targetID := strings.Repeat("a", 64), strings.Repeat("b", 64)
	storageErr := errors.New("身份仓储不可用")
	for _, test := range []struct {
		name          string
		unitErr       error
		accountErr    error
		wantErr       error
		lookupAccount bool
		grantRoles    bool
	}{
		{name: "authkit 账号直接授权", lookupAccount: true, grantRoles: true},
		{name: "外部账号不存在", accountErr: authkit.ErrNotFound, wantErr: domainrbac.ErrAccountNotFound, lookupAccount: true},
		{name: "身份仓储错误直接返回", accountErr: storageErr, wantErr: storageErr, lookupAccount: true},
		{name: "跨账号单元不查询身份", unitErr: domainunit.ErrNotFound, wantErr: domainrbac.ErrUnitNotInAccount},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &grantRepository{}
			accounts := &accountLookup{err: test.accountErr}
			units := &ownedUnitLookup{err: test.unitErr}
			service := NewService(repo, accounts, units, nil)
			ctx := appcore.SetAccountIDToContext(context.Background(), ownerID)
			_, err := service.GrantUnitRoles(
				ctx,
				&UnitAccountRolesCommand{UnitID: 7, AccountID: targetID, RoleIDs: []int{3, 3}},
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GrantUnitRoles error = %v, want %v", err, test.wantErr)
			}
			if units.accountID != ownerID || units.unitID != 7 {
				t.Fatalf("ownership lookup = %+v", units)
			}
			if (accounts.id != "") != test.lookupAccount ||
				accounts.id != "" && accounts.id != targetID {
				t.Fatalf("authkit lookup ID = %q", accounts.id)
			}
			if repo.called != test.grantRoles {
				t.Fatalf("grant repository called = %v, want %v", repo.called, test.grantRoles)
			}
			if test.grantRoles &&
				(repo.params.OwnerAccountID != ownerID || repo.params.TargetAccountID != targetID || len(repo.params.RoleIDs) != 1) {
				t.Fatalf("grant must preserve authkit IDs: %+v", repo.params)
			}
		})
	}
}

type accountLookup struct {
	id  string
	err error
}

func (a *accountLookup) GetByID(_ context.Context, id string) (*authkit.Account, error) {
	a.id = id
	if a.err != nil {
		return nil, a.err
	}
	return &authkit.Account{ID: id}, nil
}

type ownedUnitLookup struct {
	accountID string
	unitID    int
	err       error
}

func (u *ownedUnitLookup) Get(
	_ context.Context,
	accountID string,
	unitID int,
) (*domainunit.Unit, error) {
	u.accountID, u.unitID = accountID, unitID
	return &domainunit.Unit{ID: unitID, AccountID: accountID}, u.err
}

type grantRepository struct {
	domainrbac.Repository
	called bool
	params domainrbac.UnitAccountRolesParams
}

func (r *grantRepository) GrantUnitRoles(
	_ context.Context,
	params domainrbac.UnitAccountRolesParams,
) ([]*domainrbac.UnitAccountRole, error) {
	r.called, r.params = true, params
	return nil, nil
}
