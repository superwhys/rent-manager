package rbac

import "context"

type Repository interface {
	InitSystemRBAC(ctx context.Context) error
	CreateRoles(ctx context.Context, roles []*Role) ([]*Role, error)
	CreatePermissionGroups(ctx context.Context, groups []*PermissionGroup) ([]*PermissionGroup, error)
	CreatePermissions(ctx context.Context, permissions []*Permission) ([]*Permission, error)
	ListPermissionGroups(ctx context.Context) ([]*PermissionGroup, error)
	ListPermissionGroupsWithPermissions(ctx context.Context) ([]*PermissionGroup, error)
	AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error
	RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error
	AddPermissionGroupsToRole(ctx context.Context, accountID int, roleID int, permissionGroupIDs []int) error
	RemovePermissionGroupsFromRole(ctx context.Context, accountID int, roleID int, permissionGroupIDs []int) error
	AssignRolesToUnitAccount(ctx context.Context, ownerAccountID int, unitID int, accountID int, roleIDs []int) error
	HasSystemPermission(ctx context.Context, accountID int) (bool, error)
}
