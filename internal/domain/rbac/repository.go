package rbac

import "context"

type Repository interface {
	InitSystemRBAC(ctx context.Context) error
	CreateRoles(ctx context.Context, roles []*Role) ([]*Role, error)
	CreatePermissions(ctx context.Context, permissions []*Permission) ([]*Permission, error)
	AddPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
	RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error
	AssignRolesToAccount(ctx context.Context, accountID int, roleIDs []int) error
	AccountIDsByRole(ctx context.Context, roleID int) ([]int, error)
	HasPermission(ctx context.Context, accountID int, apiPath string) (bool, error)
}
