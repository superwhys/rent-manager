package rbac

import "context"

type Repository interface {
	CreateRoles(ctx context.Context, roles []*Role) ([]*Role, error)
	GetRole(ctx context.Context, accountID int, roleID int) (*Role, error)
	ListRoles(ctx context.Context, query ListRolesQuery) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, accountID int, roleID int) error
	AddPermissionGroupsToRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error
	RemovePermissionGroupsFromRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error

	CreatePermissionGroups(ctx context.Context, groups []*PermissionGroup) ([]*PermissionGroup, error)
	CreatePermissions(ctx context.Context, permissions []*Permission) ([]*Permission, error)
	ListPermissionGroups(ctx context.Context) ([]*PermissionGroup, error)
	ListPermissions(ctx context.Context) ([]*Permission, error)
	AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error
	RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error

	GrantUnitRoles(ctx context.Context, params UnitAccountRolesParams) ([]*UnitAccountRole, error)
	RevokeUnitRoles(ctx context.Context, params UnitAccountRolesParams) error
	ListUnitGrants(ctx context.Context, query ListUnitGrantsQuery) ([]*UnitAccountRole, error)
	HasUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (bool, error)
}
