package repo

import (
	"context"
	"fmt"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
	entpermission "github.com/superwhys/rent-manager-server/internal/ent/permission"
	entpermissiongroup "github.com/superwhys/rent-manager-server/internal/ent/permissiongroup"
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

	group, err := r.ensureSystemPermissionGroup(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	permission, err := r.ensureAllowAllPermission(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := r.ensurePermissionGroupHasPermission(ctx, tx, group.ID, permission.ID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *RBACRepository) ensureSystemPermissionGroup(ctx context.Context, tx *entclient.Tx) (*entclient.PermissionGroup, error) {
	group, err := tx.PermissionGroup.Query().
		Where(entpermissiongroup.Code(consts.SystemRoleSuperAdmin)).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			description := consts.SystemPermissionAllowAllDescription
			return tx.PermissionGroup.Create().
				SetCode(consts.SystemRoleSuperAdmin).
				SetName(consts.SystemRoleSuperAdmin).
				SetDescription(description).
				SetIsActive(true).
				Save(ctx)
		}
		return nil, err
	}
	return tx.PermissionGroup.UpdateOneID(group.ID).
		SetName(consts.SystemRoleSuperAdmin).
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

func (r *RBACRepository) ensurePermissionGroupHasPermission(ctx context.Context, tx *entclient.Tx, groupID int, permissionID int) error {
	exists, err := tx.PermissionGroup.Query().
		Where(entpermissiongroup.ID(groupID), entpermissiongroup.HasPermissionsWith(entpermission.ID(permissionID))).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = tx.PermissionGroup.UpdateOneID(groupID).
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
			SetOrganizationID(role.OrganizationID).
			SetName(role.Name).
			SetIsDefaultOwner(role.IsDefaultOwner).
			SetIsActive(role.IsActive),
		)
	}
	records, err := r.client.Role.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) CreatePermissionGroups(ctx context.Context, groups []*domainrbac.PermissionGroup) ([]*domainrbac.PermissionGroup, error) {
	builders := make([]*entclient.PermissionGroupCreate, 0, len(groups))
	for _, group := range groups {
		builder := r.client.PermissionGroup.Create().
			SetCode(group.Code).
			SetName(group.Name).
			SetSortOrder(group.SortOrder).
			SetIsActive(group.IsActive)
		if group.Description != nil {
			builder = builder.SetDescription(*group.Description)
		}
		builders = append(builders, builder)
	}
	records, err := r.client.PermissionGroup.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
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

func (r *RBACRepository) ListPermissionGroups(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	records, err := r.client.PermissionGroup.Query().
		Where(entpermissiongroup.DeletedAtIsNil()).
		Order(entclient.Asc(entpermissiongroup.FieldSortOrder), entclient.Desc(entpermissiongroup.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
}

func (r *RBACRepository) ListPermissionGroupsWithPermissions(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	records, err := r.client.PermissionGroup.Query().
		Where(entpermissiongroup.DeletedAtIsNil()).
		WithPermissions(func(query *entclient.PermissionQuery) {
			query.Where(entpermission.DeletedAtIsNil())
		}).
		Order(entclient.Asc(entpermissiongroup.FieldSortOrder), entclient.Desc(entpermissiongroup.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
}

func (r *RBACRepository) AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error {
	return r.client.PermissionGroup.UpdateOneID(groupID).
		AddPermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error {
	return r.client.PermissionGroup.UpdateOneID(groupID).
		RemovePermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) AddPermissionGroupsToRole(ctx context.Context, organizationID int, roleID int, permissionGroupIDs []int) error {
	return r.client.Role.Update().
		Where(entrole.ID(roleID), entrole.OrganizationID(organizationID), entrole.DeletedAtIsNil()).
		AddPermissionGroupIDs(permissionGroupIDs...).
		Exec(ctx)
}

func (r *RBACRepository) RemovePermissionGroupsFromRole(ctx context.Context, organizationID int, roleID int, permissionGroupIDs []int) error {
	return r.client.Role.Update().
		Where(entrole.ID(roleID), entrole.OrganizationID(organizationID), entrole.DeletedAtIsNil()).
		RemovePermissionGroupIDs(permissionGroupIDs...).
		Exec(ctx)
}

func (r *RBACRepository) AssignRolesToOrganizationAccount(ctx context.Context, organizationID int, accountID int, roleIDs []int) error {
	count, err := r.client.Role.Query().
		Where(entrole.IDIn(roleIDs...), entrole.OrganizationID(organizationID), entrole.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(roleIDs) {
		return domainrbac.ErrCrossOrganizationRoleAssignment
	}

	affected, err := r.client.Account.Update().
		Where(entaccount.ID(accountID), entaccount.OrganizationID(organizationID), entaccount.DeletedAtIsNil()).
		ClearRoles().
		AddRoleIDs(roleIDs...).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainrbac.ErrAccountNotInOrganization
	}
	return nil
}

func (r *RBACRepository) AccountIDsByRole(ctx context.Context, organizationID int, roleID int) ([]int, error) {
	return r.client.Account.Query().
		Where(
			entaccount.OrganizationID(organizationID),
			entaccount.DeletedAtIsNil(),
			entaccount.HasRolesWith(entrole.ID(roleID)),
		).
		IDs(ctx)
}

func (r *RBACRepository) HasPermission(ctx context.Context, accountID int, organizationID int, apiPath string) (bool, error) {
	return r.client.Account.Query().
		Where(
			entaccount.ID(accountID),
			entaccount.OrganizationID(organizationID),
			entaccount.IsActive(true),
			entaccount.DeletedAtIsNil(),
			entaccount.HasRolesWith(
				entrole.OrganizationID(organizationID),
				entrole.IsActive(true),
				entrole.DeletedAtIsNil(),
				entrole.HasPermissionGroupsWith(
					entpermissiongroup.IsActive(true),
					entpermissiongroup.DeletedAtIsNil(),
					entpermissiongroup.HasPermissionsWith(
						entpermission.Or(
							entpermission.APIPath(apiPath),
							entpermission.APIPath(consts.SystemPermissionAllowAllAPIPath),
						),
						entpermission.IsActive(true),
						entpermission.DeletedAtIsNil(),
					),
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
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		Name:           record.Name,
		IsDefaultOwner: record.IsDefaultOwner,
		IsActive:       record.IsActive,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		DeletedAt:      record.DeletedAt,
	}
}

func toDomainRoles(records []*entclient.Role) []*domainrbac.Role {
	roles := make([]*domainrbac.Role, 0, len(records))
	for _, record := range records {
		roles = append(roles, toDomainRole(record))
	}
	return roles
}

func toDomainPermissionGroup(record *entclient.PermissionGroup) *domainrbac.PermissionGroup {
	if record == nil {
		return nil
	}
	return &domainrbac.PermissionGroup{
		ID:          record.ID,
		Code:        record.Code,
		Name:        record.Name,
		Description: record.Description,
		SortOrder:   record.SortOrder,
		IsActive:    record.IsActive,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		DeletedAt:   record.DeletedAt,
	}
}

func toDomainPermissionGroups(records []*entclient.PermissionGroup) []*domainrbac.PermissionGroup {
	groups := make([]*domainrbac.PermissionGroup, 0, len(records))
	for _, record := range records {
		groups = append(groups, toDomainPermissionGroup(record))
	}
	return groups
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
