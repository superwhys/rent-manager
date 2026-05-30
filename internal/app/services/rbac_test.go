package services

import (
	"context"
	"errors"
	"testing"

	"git.superwhys.top/rent-manager/rent-manager-server/config"
	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/jwtutils"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
)

func TestRBACServiceCreatePermissionsRequiresSuperAdmin(t *testing.T) {
	repo := &fakeRBACRepository{}
	service := NewRBACService(repo, &config.AdminConfig{SuperAdminAccountIDs: []int{1}})
	cmd := &requests.CreatePermissionsCommand{
		Permissions: []requests.CreatePermissionItem{
			{Code: domainrbac.PermissionRoomView, Description: "View rooms"},
		},
	}

	_, err := service.CreatePermissions(tokenContext(2), cmd)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden for normal account, got %v", err)
	}
	if repo.createdPermissions {
		t.Fatal("normal account should not create permissions")
	}

	permissions, err := service.CreatePermissions(tokenContext(1), cmd)
	if err != nil {
		t.Fatalf("expected super admin to create permissions, got %v", err)
	}
	if len(permissions) != 1 || permissions[0].Code != domainrbac.PermissionRoomView {
		t.Fatalf("unexpected permissions: %#v", permissions)
	}
}

func TestRBACServiceRequireUnitPermission(t *testing.T) {
	service := NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {
				10: {
					domainrbac.PermissionRoomView: true,
				},
			},
		},
	}, nil)

	if err := service.RequireUnitPermission(context.Background(), 2, 10, domainrbac.PermissionRoomView); err != nil {
		t.Fatalf("expected granted permission, got %v", err)
	}

	err := service.RequireUnitPermission(context.Background(), 2, 10, domainrbac.PermissionRoomManage)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden for missing permission, got %v", err)
	}

	err = service.RequireUnitPermission(context.Background(), 2, 99, domainrbac.PermissionRoomView)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden for ungranted unit, got %v", err)
	}
}

func tokenContext(accountID int) context.Context {
	return jwtutils.SetTokenClaimsToContext(context.Background(), &jwtutils.AccountTokenClaims{
		UserID:    accountID,
		TokenType: jwtutils.TokenTypeAccess,
	})
}

type fakeRBACRepository struct {
	createdPermissions bool
	unitPermissions    map[int]map[int]map[string]bool
}

func (f *fakeRBACRepository) CreateRoles(context.Context, []*domainrbac.Role) ([]*domainrbac.Role, error) {
	return nil, nil
}

func (f *fakeRBACRepository) GetRole(context.Context, int, int) (*domainrbac.Role, error) {
	return nil, domainrbac.ErrRoleNotFound
}

func (f *fakeRBACRepository) ListRoles(context.Context, domainrbac.ListRolesQuery) ([]*domainrbac.Role, error) {
	return nil, nil
}

func (f *fakeRBACRepository) UpdateRole(context.Context, *domainrbac.Role) (*domainrbac.Role, error) {
	return nil, nil
}

func (f *fakeRBACRepository) DeleteRole(context.Context, int, int) error {
	return nil
}

func (f *fakeRBACRepository) AddPermissionGroupsToRole(context.Context, int, int, []int) error {
	return nil
}

func (f *fakeRBACRepository) RemovePermissionGroupsFromRole(context.Context, int, int, []int) error {
	return nil
}

func (f *fakeRBACRepository) CreatePermissionGroups(context.Context, []*domainrbac.PermissionGroup) ([]*domainrbac.PermissionGroup, error) {
	return nil, nil
}

func (f *fakeRBACRepository) CreatePermissions(_ context.Context, permissions []*domainrbac.Permission) ([]*domainrbac.Permission, error) {
	f.createdPermissions = true
	return permissions, nil
}

func (f *fakeRBACRepository) ListPermissionGroups(context.Context) ([]*domainrbac.PermissionGroup, error) {
	return nil, nil
}

func (f *fakeRBACRepository) ListPermissions(context.Context) ([]*domainrbac.Permission, error) {
	return nil, nil
}

func (f *fakeRBACRepository) AddPermissionsToPermissionGroup(context.Context, int, []int) error {
	return nil
}

func (f *fakeRBACRepository) RemovePermissionsFromPermissionGroup(context.Context, int, []int) error {
	return nil
}

func (f *fakeRBACRepository) GrantUnitRoles(context.Context, domainrbac.UnitAccountRolesParams) ([]*domainrbac.UnitAccountRole, error) {
	return nil, nil
}

func (f *fakeRBACRepository) RevokeUnitRoles(context.Context, domainrbac.UnitAccountRolesParams) error {
	return nil
}

func (f *fakeRBACRepository) ListUnitGrants(context.Context, domainrbac.ListUnitGrantsQuery) ([]*domainrbac.UnitAccountRole, error) {
	return nil, nil
}

func (f *fakeRBACRepository) HasUnitPermission(_ context.Context, accountID int, unitID int, permissionCode string) (bool, error) {
	if f.unitPermissions == nil {
		return false, nil
	}
	unitPermissions, ok := f.unitPermissions[accountID]
	if !ok {
		return false, nil
	}
	permissions, ok := unitPermissions[unitID]
	if !ok {
		return false, nil
	}
	return permissions[permissionCode], nil
}
