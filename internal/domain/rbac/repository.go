package rbac

import "context"

type Repository interface {
	CreatePermissionGroups(ctx context.Context, groups []*PermissionGroup) ([]*PermissionGroup, error)
	CreatePermissions(ctx context.Context, permissions []*Permission) ([]*Permission, error)
	ListPermissionGroups(ctx context.Context) ([]*PermissionGroup, error)
	ListPermissions(ctx context.Context) ([]*Permission, error)
	AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error
	RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error
}
