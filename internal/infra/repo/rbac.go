package repo

import (
	"context"
	"fmt"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
	entpermission "github.com/superwhys/rent-manager-server/internal/ent/permission"
	entrole "github.com/superwhys/rent-manager-server/internal/ent/role"
	"github.com/superwhys/rent-manager-server/internal/models/consts"
)

type RBACRepository struct {
	client *entclient.Client
}

func NewRBACRepository(client *entclient.Client) *RBACRepository {
	return &RBACRepository{client: client}
}

func (r *RBACRepository) InitSystemRBAC(ctx context.Context) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}

	role, err := r.ensureSuperAdminRole(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	permission, err := r.ensureAllowAllPermission(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := r.ensureRoleHasPermission(ctx, tx, role.ID, permission.ID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *RBACRepository) ensureSuperAdminRole(ctx context.Context, tx *entclient.Tx) (*entclient.Role, error) {
	role, err := tx.Role.Query().
		Where(entrole.Name(consts.SystemRoleSuperAdmin)).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return tx.Role.Create().
				SetName(consts.SystemRoleSuperAdmin).
				SetIsActive(true).
				Save(ctx)
		}
		return nil, err
	}
	return tx.Role.UpdateOneID(role.ID).
		SetIsActive(true).
		ClearDeletedAt().
		Save(ctx)
}

func (r *RBACRepository) ensureAllowAllPermission(ctx context.Context, tx *entclient.Tx) (*entclient.Permission, error) {
	permissions, err := tx.Permission.Query().
		Where(
			entpermission.Or(
				entpermission.Code(consts.SystemPermissionAllowAllCode),
				entpermission.APIPath(consts.SystemPermissionAllowAllAPIPath),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return tx.Permission.Create().
			SetCode(consts.SystemPermissionAllowAllCode).
			SetAPIPath(consts.SystemPermissionAllowAllAPIPath).
			SetDescription(consts.SystemPermissionAllowAllDescription).
			SetIsActive(true).
			Save(ctx)
	}
	if len(permissions) > 1 {
		return nil, fmt.Errorf("conflicting system allow-all permissions")
	}

	permission := permissions[0]
	if permission.Code != consts.SystemPermissionAllowAllCode || permission.APIPath != consts.SystemPermissionAllowAllAPIPath {
		return nil, fmt.Errorf("conflicting system allow-all permission: id=%d code=%q api_path=%q", permission.ID, permission.Code, permission.APIPath)
	}

	return tx.Permission.UpdateOneID(permission.ID).
		SetDescription(consts.SystemPermissionAllowAllDescription).
		SetIsActive(true).
		ClearDeletedAt().
		Save(ctx)
}

func (r *RBACRepository) ensureRoleHasPermission(ctx context.Context, tx *entclient.Tx, roleID int, permissionID int) error {
	exists, err := tx.Role.Query().
		Where(
			entrole.ID(roleID),
			entrole.HasPermissionsWith(entpermission.ID(permissionID)),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = tx.Role.UpdateOneID(roleID).
		AddPermissionIDs(permissionID).
		Exec(ctx)
	if entclient.IsConstraintError(err) {
		return nil
	}
	return err
}

func (r *RBACRepository) CreateRoles(ctx context.Context, roles []*domainrbac.Role) ([]*domainrbac.Role, error) {
	builders := make([]*entclient.RoleCreate, 0, len(roles))
	for _, role := range roles {
		builders = append(builders, r.client.Role.Create().
			SetName(role.Name).
			SetIsActive(role.IsActive),
		)
	}
	records, err := r.client.Role.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) CreatePermissions(ctx context.Context, permissions []*domainrbac.Permission) ([]*domainrbac.Permission, error) {
	builders := make([]*entclient.PermissionCreate, 0, len(permissions))
	for _, permission := range permissions {
		builders = append(builders, r.client.Permission.Create().
			SetCode(permission.Code).
			SetAPIPath(permission.APIPath).
			SetDescription(permission.Description).
			SetIsActive(permission.IsActive),
		)
	}
	records, err := r.client.Permission.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPermissions(records), nil
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

func (r *RBACRepository) AccountIDsByRole(ctx context.Context, roleID int) ([]int, error) {
	return r.client.Account.Query().
		Where(
			entaccount.DeletedAtIsNil(),
			entaccount.HasRolesWith(entrole.ID(roleID)),
		).
		IDs(ctx)
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
					entpermission.Or(
						entpermission.APIPath(apiPath),
						entpermission.APIPath(consts.SystemPermissionAllowAllAPIPath),
					),
					entpermission.IsActive(true),
					entpermission.DeletedAtIsNil(),
				),
			),
		).
		Exist(ctx)
}

func toDomainRole(record *entclient.Role) *domainrbac.Role {
	if record == nil {
		return nil
	}
	return &domainrbac.Role{
		ID:        record.ID,
		Name:      record.Name,
		IsActive:  record.IsActive,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}

func toDomainRoles(records []*entclient.Role) []*domainrbac.Role {
	roles := make([]*domainrbac.Role, 0, len(records))
	for _, record := range records {
		roles = append(roles, toDomainRole(record))
	}
	return roles
}

func toDomainPermission(record *entclient.Permission) *domainrbac.Permission {
	if record == nil {
		return nil
	}
	return &domainrbac.Permission{
		ID:          record.ID,
		Code:        record.Code,
		APIPath:     record.APIPath,
		Description: record.Description,
		IsActive:    record.IsActive,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		DeletedAt:   record.DeletedAt,
	}
}

func toDomainPermissions(records []*entclient.Permission) []*domainrbac.Permission {
	permissions := make([]*domainrbac.Permission, 0, len(records))
	for _, record := range records {
		permissions = append(permissions, toDomainPermission(record))
	}
	return permissions
}
