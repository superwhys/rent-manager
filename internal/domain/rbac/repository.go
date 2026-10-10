package rbac

import "context"

type Repository interface {
	CreateRoles(ctx context.Context, roles []*Role) ([]*Role, error)
	GetRole(ctx context.Context, accountID string, roleID int) (*Role, error)
	ListRoles(ctx context.Context, query ListRolesQuery) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, accountID string, roleID int) error
	AddPermissionGroupsToRole(
		ctx context.Context,
		accountID string,
		roleID int,
		groupIDs []int,
	) error
	RemovePermissionGroupsFromRole(
		ctx context.Context,
		accountID string,
		roleID int,
		groupIDs []int,
	) error

	CreatePermissionGroups(
		ctx context.Context,
		groups []*PermissionGroup,
	) ([]*PermissionGroup, error)
	ListPermissionGroups(ctx context.Context) ([]*PermissionGroup, error)
	AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionCodes []string) error
	RemovePermissionsFromPermissionGroup(
		ctx context.Context,
		groupID int,
		permissionCodes []string,
	) error

	GrantUnitRoles(ctx context.Context, params UnitAccountRolesParams) ([]*UnitAccountRole, error)
	RevokeUnitRoles(ctx context.Context, params UnitAccountRolesParams) error
	ListUnitGrants(ctx context.Context, query ListUnitGrantsQuery) ([]*UnitAccountRole, error)
	HasUnitPermission(
		ctx context.Context,
		accountID string,
		unitID int,
		permissionCode string,
	) (bool, error)
}
