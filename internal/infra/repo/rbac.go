package repo

import (
	"context"

	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
	entpermission "github.com/superwhys/rent-manager-server/internal/ent/permission"
	entrole "github.com/superwhys/rent-manager-server/internal/ent/role"
)

type CreateRoleParams struct {
	Name     string
	IsActive bool
}

type CreatePermissionParams struct {
	Code        string
	APIPath     string
	Description string
	IsActive    bool
}

type RBACRepository struct {
	client *entclient.Client
}

func NewRBACRepository(client *entclient.Client) *RBACRepository {
	return &RBACRepository{client: client}
}

func (r *RBACRepository) CreateRoles(ctx context.Context, params []CreateRoleParams) ([]*entclient.Role, error) {
	builders := make([]*entclient.RoleCreate, 0, len(params))
	for _, param := range params {
		builders = append(builders, r.client.Role.Create().
			SetName(param.Name).
			SetIsActive(param.IsActive),
		)
	}
	return r.client.Role.CreateBulk(builders...).Save(ctx)
}

func (r *RBACRepository) CreatePermissions(ctx context.Context, params []CreatePermissionParams) ([]*entclient.Permission, error) {
	builders := make([]*entclient.PermissionCreate, 0, len(params))
	for _, param := range params {
		builders = append(builders, r.client.Permission.Create().
			SetCode(param.Code).
			SetAPIPath(param.APIPath).
			SetDescription(param.Description).
			SetIsActive(param.IsActive),
		)
	}
	return r.client.Permission.CreateBulk(builders...).Save(ctx)
}

func (r *RBACRepository) AddPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	return r.client.Role.UpdateOneID(roleID).
		AddPermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error {
	return r.client.Role.UpdateOneID(roleID).
		RemovePermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) AssignRolesToAccount(ctx context.Context, accountID int, roleIDs []int) error {
	return r.client.Account.UpdateOneID(accountID).
		AddRoleIDs(roleIDs...).
		Exec(ctx)
}

func (r *RBACRepository) HasPermission(ctx context.Context, accountID int, apiPath string) (bool, error) {
	return r.client.Account.Query().
		Where(
			entaccount.ID(accountID),
			entaccount.IsActive(true),
			entaccount.DeletedAtIsNil(),
			entaccount.HasRolesWith(
				entrole.IsActive(true),
				entrole.DeletedAtIsNil(),
				entrole.HasPermissionsWith(
					entpermission.APIPath(apiPath),
					entpermission.IsActive(true),
					entpermission.DeletedAtIsNil(),
				),
			),
		).
		Exist(ctx)
}

func (r *RBACRepository) HasAccountRoles(ctx context.Context) (bool, error) {
	return r.client.Account.Query().
		Where(
			entaccount.DeletedAtIsNil(),
			entaccount.HasRoles(),
		).
		Exist(ctx)
}
