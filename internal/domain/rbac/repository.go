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
	AddPermissionGroupsToRole(ctx context.Context, organizationID int, roleID int, permissionGroupIDs []int) error
	RemovePermissionGroupsFromRole(ctx context.Context, organizationID int, roleID int, permissionGroupIDs []int) error
	AssignRolesToOrganizationAccount(ctx context.Context, organizationID int, accountID int, roleIDs []int) error
	AccountIDsByRole(ctx context.Context, organizationID int, roleID int) ([]int, error)
	HasPermission(ctx context.Context, accountID int, organizationID int, apiPath string) (bool, error)
}
