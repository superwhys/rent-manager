package repo

import (
	"context"
	"time"

	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	entclient "git.superwhys.top/rent-manager/rent-manager-server/internal/ent"
	entaccount "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/account"
	entpermission "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/permission"
	entpermissiongroup "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/permissiongroup"
	entrole "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/role"
	entunit "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/unit"
	entunitaccountrole "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/unitaccountrole"
)

type RBACRepository struct {
	client *entclient.Client
}

func NewRBACRepository(client *entclient.Client) *RBACRepository {
	return &RBACRepository{client: client}
}

func (r *RBACRepository) CreateRoles(ctx context.Context, roles []*domainrbac.Role) ([]*domainrbac.Role, error) {
	builders := make([]*entclient.RoleCreate, 0, len(roles))
	for _, role := range roles {
		builders = append(builders, r.client.Role.Create().
			SetAccountID(role.AccountID).
			SetName(role.Name).
			SetIsDefaultOwner(role.IsDefaultOwner).
			SetIsActive(role.IsActive),
		)
	}
	records, err := r.client.Role.CreateBulk(builders...).Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainrbac.ErrRoleAlreadyExists
		}
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) GetRole(ctx context.Context, accountID int, roleID int) (*domainrbac.Role, error) {
	record, err := r.client.Role.Query().
		Where(entrole.ID(roleID), entrole.AccountID(accountID), entrole.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainrbac.ErrRoleNotFound
		}
		return nil, err
	}
	return toDomainRole(record), nil
}

func (r *RBACRepository) ListRoles(ctx context.Context, query domainrbac.ListRolesQuery) ([]*domainrbac.Role, error) {
	q := r.client.Role.Query().
		Where(entrole.AccountID(query.AccountID), entrole.DeletedAtIsNil())
	if query.IsActive != nil {
		q = q.Where(entrole.IsActive(*query.IsActive))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entrole.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) UpdateRole(ctx context.Context, role *domainrbac.Role) (*domainrbac.Role, error) {
	affected, err := r.client.Role.Update().
		Where(entrole.ID(role.ID), entrole.AccountID(role.AccountID), entrole.DeletedAtIsNil()).
		SetName(role.Name).
		SetIsDefaultOwner(role.IsDefaultOwner).
		SetIsActive(role.IsActive).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainrbac.ErrRoleAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainrbac.ErrRoleNotFound
	}
	return r.GetRole(ctx, role.AccountID, role.ID)
}

func (r *RBACRepository) DeleteRole(ctx context.Context, accountID int, roleID int) error {
	now := time.Now()
	affected, err := r.client.Role.Update().
		Where(entrole.ID(roleID), entrole.AccountID(accountID), entrole.DeletedAtIsNil()).
		SetDeletedAt(now).
		SetIsActive(false).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainrbac.ErrRoleNotFound
	}
	_, err = r.client.UnitAccountRole.Update().
		Where(entunitaccountrole.RoleID(roleID), entunitaccountrole.DeletedAtIsNil()).
		SetStatus(entunitaccountrole.StatusDisabled).
		Save(ctx)
	return err
}

func (r *RBACRepository) AddPermissionGroupsToRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error {
	if err := r.ensureRoleInAccount(ctx, accountID, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionGroupsExist(ctx, groupIDs); err != nil {
		return err
	}
	return r.client.Role.UpdateOneID(roleID).AddPermissionGroupIDs(groupIDs...).Exec(ctx)
}

func (r *RBACRepository) RemovePermissionGroupsFromRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error {
	if err := r.ensureRoleInAccount(ctx, accountID, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionGroupsExist(ctx, groupIDs); err != nil {
		return err
	}
	return r.client.Role.UpdateOneID(roleID).RemovePermissionGroupIDs(groupIDs...).Exec(ctx)
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
		if entclient.IsConstraintError(err) {
			return nil, domainrbac.ErrPermissionGroupAlreadyExists
		}
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
}

func (r *RBACRepository) CreatePermissions(ctx context.Context, permissions []*domainrbac.Permission) ([]*domainrbac.Permission, error) {
	builders := make([]*entclient.PermissionCreate, 0, len(permissions))
	for _, permission := range permissions {
		builders = append(builders, r.client.Permission.Create().
			SetCode(permission.Code).
			SetDescription(permission.Description).
			SetIsActive(permission.IsActive),
		)
	}
	records, err := r.client.Permission.CreateBulk(builders...).Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainrbac.ErrPermissionAlreadyExists
		}
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

func (r *RBACRepository) ListPermissions(ctx context.Context) ([]*domainrbac.Permission, error) {
	records, err := r.client.Permission.Query().
		Where(entpermission.DeletedAtIsNil()).
		Order(entclient.Desc(entpermission.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainPermissions(records), nil
}

func (r *RBACRepository) AddPermissionsToPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error {
	if err := r.ensurePermissionGroupExists(ctx, groupID); err != nil {
		return err
	}
	if err := r.ensurePermissionsExist(ctx, permissionIDs); err != nil {
		return err
	}
	return r.client.PermissionGroup.UpdateOneID(groupID).
		AddPermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error {
	if err := r.ensurePermissionGroupExists(ctx, groupID); err != nil {
		return err
	}
	if err := r.ensurePermissionsExist(ctx, permissionIDs); err != nil {
		return err
	}
	return r.client.PermissionGroup.UpdateOneID(groupID).
		RemovePermissionIDs(permissionIDs...).
		Exec(ctx)
}

func (r *RBACRepository) GrantUnitRoles(ctx context.Context, params domainrbac.UnitAccountRolesParams) ([]*domainrbac.UnitAccountRole, error) {
	if err := r.ensureUnitOwnedByAccount(ctx, params.OwnerAccountID, params.UnitID); err != nil {
		return nil, err
	}
	if err := r.ensureAccountExists(ctx, params.TargetAccountID); err != nil {
		return nil, err
	}
	if err := r.ensureRolesInAccount(ctx, params.OwnerAccountID, params.RoleIDs); err != nil {
		return nil, err
	}

	for _, roleID := range params.RoleIDs {
		record, err := r.client.UnitAccountRole.Query().
			Where(
				entunitaccountrole.UnitID(params.UnitID),
				entunitaccountrole.AccountID(params.TargetAccountID),
				entunitaccountrole.RoleID(roleID),
				entunitaccountrole.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			if !entclient.IsNotFound(err) {
				return nil, err
			}
			_, err = r.client.UnitAccountRole.Create().
				SetUnitID(params.UnitID).
				SetAccountID(params.TargetAccountID).
				SetRoleID(roleID).
				SetGrantedByAccountID(params.OwnerAccountID).
				SetStatus(entunitaccountrole.StatusActive).
				Save(ctx)
			if err != nil {
				if entclient.IsConstraintError(err) {
					return nil, domainrbac.ErrUnitGrantNotFound
				}
				return nil, err
			}
			continue
		}
		if record.GrantedByAccountID != params.OwnerAccountID {
			return nil, domainrbac.ErrUnitNotInAccount
		}
		_, err = r.client.UnitAccountRole.UpdateOneID(record.ID).
			SetStatus(entunitaccountrole.StatusActive).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	return r.ListUnitGrants(ctx, domainrbac.ListUnitGrantsQuery{
		OwnerAccountID:  params.OwnerAccountID,
		UnitID:          params.UnitID,
		TargetAccountID: params.TargetAccountID,
		Status:          domainrbac.UnitGrantStatusActive,
	})
}

func (r *RBACRepository) RevokeUnitRoles(ctx context.Context, params domainrbac.UnitAccountRolesParams) error {
	if err := r.ensureUnitOwnedByAccount(ctx, params.OwnerAccountID, params.UnitID); err != nil {
		return err
	}
	affected, err := r.client.UnitAccountRole.Update().
		Where(
			entunitaccountrole.UnitID(params.UnitID),
			entunitaccountrole.AccountID(params.TargetAccountID),
			entunitaccountrole.RoleIDIn(params.RoleIDs...),
			entunitaccountrole.GrantedByAccountID(params.OwnerAccountID),
			entunitaccountrole.DeletedAtIsNil(),
		).
		SetStatus(entunitaccountrole.StatusDisabled).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainrbac.ErrUnitGrantNotFound
	}
	return nil
}

func (r *RBACRepository) ListUnitGrants(ctx context.Context, query domainrbac.ListUnitGrantsQuery) ([]*domainrbac.UnitAccountRole, error) {
	if err := r.ensureUnitOwnedByAccount(ctx, query.OwnerAccountID, query.UnitID); err != nil {
		return nil, err
	}
	q := r.client.UnitAccountRole.Query().
		Where(
			entunitaccountrole.UnitID(query.UnitID),
			entunitaccountrole.GrantedByAccountID(query.OwnerAccountID),
			entunitaccountrole.DeletedAtIsNil(),
		).
		WithRole()
	if query.TargetAccountID > 0 {
		q = q.Where(entunitaccountrole.AccountID(query.TargetAccountID))
	}
	if query.Status != "" {
		q = q.Where(entunitaccountrole.StatusEQ(entunitaccountrole.Status(query.Status)))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entunitaccountrole.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainUnitAccountRoles(records), nil
}

func (r *RBACRepository) HasUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (bool, error) {
	unitRecord, err := r.client.Unit.Query().
		Where(entunit.ID(unitID), entunit.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return false, domainrbac.ErrUnitNotInAccount
		}
		return false, err
	}
	if unitRecord.AccountID == accountID {
		return true, nil
	}

	return r.client.UnitAccountRole.Query().
		Where(
			entunitaccountrole.UnitID(unitID),
			entunitaccountrole.AccountID(accountID),
			entunitaccountrole.StatusEQ(entunitaccountrole.StatusActive),
			entunitaccountrole.DeletedAtIsNil(),
			entunitaccountrole.HasRoleWith(
				entrole.IsActive(true),
				entrole.DeletedAtIsNil(),
				entrole.HasPermissionGroupsWith(
					entpermissiongroup.IsActive(true),
					entpermissiongroup.DeletedAtIsNil(),
					entpermissiongroup.HasPermissionsWith(
						entpermission.Code(permissionCode),
						entpermission.IsActive(true),
						entpermission.DeletedAtIsNil(),
					),
				),
			),
		).
		Exist(ctx)
}

func (r *RBACRepository) ensureRoleInAccount(ctx context.Context, accountID int, roleID int) error {
	exists, err := r.client.Role.Query().
		Where(entrole.ID(roleID), entrole.AccountID(accountID), entrole.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domainrbac.ErrRoleNotFound
	}
	return nil
}

func (r *RBACRepository) ensureRolesInAccount(ctx context.Context, accountID int, roleIDs []int) error {
	count, err := r.client.Role.Query().
		Where(entrole.IDIn(roleIDs...), entrole.AccountID(accountID), entrole.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(roleIDs) {
		return domainrbac.ErrRoleNotInAccount
	}
	return nil
}

func (r *RBACRepository) ensureUnitOwnedByAccount(ctx context.Context, accountID int, unitID int) error {
	exists, err := r.client.Unit.Query().
		Where(entunit.ID(unitID), entunit.AccountID(accountID), entunit.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domainrbac.ErrUnitNotInAccount
	}
	return nil
}

func (r *RBACRepository) ensureAccountExists(ctx context.Context, accountID int) error {
	exists, err := r.client.Account.Query().
		Where(entaccount.ID(accountID), entaccount.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domainrbac.ErrAccountNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionGroupExists(ctx context.Context, groupID int) error {
	exists, err := r.client.PermissionGroup.Query().
		Where(entpermissiongroup.ID(groupID), entpermissiongroup.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domainrbac.ErrPermissionGroupNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionGroupsExist(ctx context.Context, groupIDs []int) error {
	count, err := r.client.PermissionGroup.Query().
		Where(entpermissiongroup.IDIn(groupIDs...), entpermissiongroup.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(groupIDs) {
		return domainrbac.ErrPermissionGroupNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionsExist(ctx context.Context, permissionIDs []int) error {
	count, err := r.client.Permission.Query().
		Where(entpermission.IDIn(permissionIDs...), entpermission.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(permissionIDs) {
		return domainrbac.ErrPermissionNotFound
	}
	return nil
}

func toDomainRole(record *entclient.Role) *domainrbac.Role {
	if record == nil {
		return nil
	}
	return &domainrbac.Role{
		ID:             record.ID,
		AccountID:      record.AccountID,
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

func toDomainUnitAccountRole(record *entclient.UnitAccountRole) *domainrbac.UnitAccountRole {
	if record == nil {
		return nil
	}
	grant := &domainrbac.UnitAccountRole{
		ID:                 record.ID,
		UnitID:             record.UnitID,
		AccountID:          record.AccountID,
		RoleID:             record.RoleID,
		GrantedByAccountID: record.GrantedByAccountID,
		Status:             record.Status.String(),
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
		DeletedAt:          record.DeletedAt,
	}
	if record.Edges.Role != nil {
		grant.Role = toDomainRole(record.Edges.Role)
	}
	return grant
}

func toDomainUnitAccountRoles(records []*entclient.UnitAccountRole) []*domainrbac.UnitAccountRole {
	grants := make([]*domainrbac.UnitAccountRole, 0, len(records))
	for _, record := range records {
		grants = append(grants, toDomainUnitAccountRole(record))
	}
	return grants
}
