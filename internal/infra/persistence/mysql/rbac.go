package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type RBACRepository struct {
	q *query.Query
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{q: query.Use(db)}
}

func (r *RBACRepository) CreateRoles(ctx context.Context, roles []*domainrbac.Role) ([]*domainrbac.Role, error) {
	records := make([]*models.Role, 0, len(roles))
	for _, role := range roles {
		if err := validateByteLength(255, role.Name); err != nil {
			return nil, err
		}
		records = append(records, &models.Role{
			AccountID:      role.AccountID,
			Name:           role.Name,
			IsDefaultOwner: role.IsDefaultOwner,
			IsActive:       &role.IsActive,
		})
	}
	if err := r.q.Role.WithContext(ctx).Create(records...); err != nil {
		if isConstraintError(err) {
			return nil, domainrbac.ErrRoleAlreadyExists
		}
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) GetRole(ctx context.Context, accountID int, roleID int) (*domainrbac.Role, error) {
	role := r.q.Role
	record, err := role.WithContext(ctx).
		Where(role.ID.Eq(roleID), role.AccountID.Eq(accountID), role.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainrbac.ErrRoleNotFound
		}
		return nil, err
	}
	return toDomainRole(record), nil
}

func (r *RBACRepository) ListRoles(ctx context.Context, query domainrbac.ListRolesQuery) ([]*domainrbac.Role, error) {
	role := r.q.Role
	q := role.WithContext(ctx).
		Where(role.AccountID.Eq(query.AccountID), role.DeletedAt.IsNull())
	if query.IsActive != nil {
		q = q.Where(role.IsActive.Is(*query.IsActive))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(role.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainRoles(records), nil
}

func (r *RBACRepository) UpdateRole(ctx context.Context, role *domainrbac.Role) (*domainrbac.Role, error) {
	if err := validateByteLength(255, role.Name); err != nil {
		return nil, err
	}
	record := r.q.Role
	result, err := record.WithContext(ctx).
		Where(record.ID.Eq(role.ID), record.AccountID.Eq(role.AccountID), record.DeletedAt.IsNull()).
		UpdateSimple(
			record.Name.Value(role.Name),
			record.IsDefaultOwner.Value(role.IsDefaultOwner),
			record.IsActive.Value(role.IsActive),
		)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainrbac.ErrRoleAlreadyExists
		}
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, domainrbac.ErrRoleNotFound
	}
	return r.GetRole(ctx, role.AccountID, role.ID)
}

func (r *RBACRepository) DeleteRole(ctx context.Context, accountID int, roleID int) error {
	role := r.q.Role
	result, err := role.WithContext(ctx).
		Where(role.ID.Eq(roleID), role.AccountID.Eq(accountID), role.DeletedAt.IsNull()).
		UpdateSimple(role.DeletedAt.Value(time.Now()), role.IsActive.Value(false))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainrbac.ErrRoleNotFound
	}
	grant := r.q.UnitAccountRole
	_, err = grant.WithContext(ctx).
		Where(grant.RoleID.Eq(roleID), grant.DeletedAt.IsNull()).
		UpdateSimple(grant.Status.Value(domainrbac.UnitGrantStatusDisabled))
	return err
}

func (r *RBACRepository) AddPermissionGroupsToRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error {
	if err := r.ensureRoleInAccount(ctx, accountID, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionGroupsExist(ctx, groupIDs); err != nil {
		return err
	}
	return query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		records := make([]*models.RolePermissionGroup, 0, len(groupIDs))
		for _, groupID := range groupIDs {
			records = append(records, &models.RolePermissionGroup{RoleID: roleID, PermissionGroupID: groupID})
		}
		if err := tx.RolePermissionGroup.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(records...); err != nil {
			return err
		}
		_, err := tx.Role.WithContext(ctx).Where(tx.Role.ID.Eq(roleID)).UpdateSimple(tx.Role.UpdatedAt.Value(time.Now()))
		return err
	})
}

func (r *RBACRepository) RemovePermissionGroupsFromRole(ctx context.Context, accountID int, roleID int, groupIDs []int) error {
	if err := r.ensureRoleInAccount(ctx, accountID, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionGroupsExist(ctx, groupIDs); err != nil {
		return err
	}
	return query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		link := tx.RolePermissionGroup
		if _, err := link.WithContext(ctx).Where(link.RoleID.Eq(roleID), link.PermissionGroupID.In(groupIDs...)).Delete(); err != nil {
			return err
		}
		_, err := tx.Role.WithContext(ctx).Where(tx.Role.ID.Eq(roleID)).UpdateSimple(tx.Role.UpdatedAt.Value(time.Now()))
		return err
	})
}

func (r *RBACRepository) CreatePermissionGroups(ctx context.Context, groups []*domainrbac.PermissionGroup) ([]*domainrbac.PermissionGroup, error) {
	records := make([]*models.PermissionGroup, 0, len(groups))
	for _, group := range groups {
		if err := validateByteLength(255, group.Code, group.Name); err != nil {
			return nil, err
		}
		if group.Description != nil {
			if err := validateByteLength(255, *group.Description); err != nil {
				return nil, err
			}
		}
		records = append(records, &models.PermissionGroup{
			Code:        group.Code,
			Name:        group.Name,
			Description: group.Description,
			SortOrder:   group.SortOrder,
			IsActive:    &group.IsActive,
		})
	}
	if err := r.q.PermissionGroup.WithContext(ctx).Create(records...); err != nil {
		if isConstraintError(err) {
			return nil, domainrbac.ErrPermissionGroupAlreadyExists
		}
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
}

func (r *RBACRepository) CreatePermissions(ctx context.Context, permissions []*domainrbac.Permission) ([]*domainrbac.Permission, error) {
	records := make([]*models.Permission, 0, len(permissions))
	for _, permission := range permissions {
		if err := validateByteLength(255, permission.Code, permission.Description); err != nil {
			return nil, err
		}
		records = append(records, &models.Permission{
			Code:        permission.Code,
			Description: permission.Description,
			IsActive:    &permission.IsActive,
		})
	}
	if err := r.q.Permission.WithContext(ctx).Create(records...); err != nil {
		if isConstraintError(err) {
			return nil, domainrbac.ErrPermissionAlreadyExists
		}
		return nil, err
	}
	return toDomainPermissions(records), nil
}

func (r *RBACRepository) ListPermissionGroups(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	group := r.q.PermissionGroup
	records, err := group.WithContext(ctx).
		Where(group.DeletedAt.IsNull()).
		Order(group.SortOrder.Asc(), group.ID.Desc()).
		Find()
	if err != nil {
		return nil, err
	}
	return toDomainPermissionGroups(records), nil
}

func (r *RBACRepository) ListPermissions(ctx context.Context) ([]*domainrbac.Permission, error) {
	permission := r.q.Permission
	records, err := permission.WithContext(ctx).
		Where(permission.DeletedAt.IsNull()).
		Order(permission.ID.Desc()).
		Find()
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
	return query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		records := make([]*models.PermissionGroupPermission, 0, len(permissionIDs))
		for _, permissionID := range permissionIDs {
			records = append(records, &models.PermissionGroupPermission{PermissionGroupID: groupID, PermissionID: permissionID})
		}
		if err := tx.PermissionGroupPermission.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(records...); err != nil {
			return err
		}
		_, err := tx.PermissionGroup.WithContext(ctx).Where(tx.PermissionGroup.ID.Eq(groupID)).UpdateSimple(tx.PermissionGroup.UpdatedAt.Value(time.Now()))
		return err
	})
}

func (r *RBACRepository) RemovePermissionsFromPermissionGroup(ctx context.Context, groupID int, permissionIDs []int) error {
	if err := r.ensurePermissionGroupExists(ctx, groupID); err != nil {
		return err
	}
	if err := r.ensurePermissionsExist(ctx, permissionIDs); err != nil {
		return err
	}
	return query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		link := tx.PermissionGroupPermission
		if _, err := link.WithContext(ctx).Where(link.PermissionGroupID.Eq(groupID), link.PermissionID.In(permissionIDs...)).Delete(); err != nil {
			return err
		}
		_, err := tx.PermissionGroup.WithContext(ctx).Where(tx.PermissionGroup.ID.Eq(groupID)).UpdateSimple(tx.PermissionGroup.UpdatedAt.Value(time.Now()))
		return err
	})
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

	grant := r.q.UnitAccountRole
	for _, roleID := range params.RoleIDs {
		record, err := grant.WithContext(ctx).
			Where(
				grant.UnitID.Eq(params.UnitID),
				grant.AccountID.Eq(params.TargetAccountID),
				grant.RoleID.Eq(roleID),
				grant.DeletedAt.IsNull(),
			).
			Take()
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			err = grant.WithContext(ctx).Create(&models.UnitAccountRole{
				UnitID:             params.UnitID,
				AccountID:          params.TargetAccountID,
				RoleID:             roleID,
				GrantedByAccountID: params.OwnerAccountID,
				Status:             domainrbac.UnitGrantStatusActive,
			})
			if err != nil {
				if isConstraintError(err) {
					return nil, domainrbac.ErrUnitGrantNotFound
				}
				return nil, err
			}
			continue
		}
		if record.GrantedByAccountID != params.OwnerAccountID {
			return nil, domainrbac.ErrUnitNotInAccount
		}
		_, err = grant.WithContext(ctx).Where(grant.ID.Eq(record.ID)).UpdateSimple(grant.Status.Value(domainrbac.UnitGrantStatusActive))
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
	grant := r.q.UnitAccountRole
	result, err := grant.WithContext(ctx).
		Where(
			grant.UnitID.Eq(params.UnitID),
			grant.AccountID.Eq(params.TargetAccountID),
			grant.RoleID.In(params.RoleIDs...),
			grant.GrantedByAccountID.Eq(params.OwnerAccountID),
			grant.DeletedAt.IsNull(),
		).
		UpdateSimple(grant.Status.Value(domainrbac.UnitGrantStatusDisabled))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainrbac.ErrUnitGrantNotFound
	}
	return nil
}

func (r *RBACRepository) ListUnitGrants(ctx context.Context, query domainrbac.ListUnitGrantsQuery) ([]*domainrbac.UnitAccountRole, error) {
	if err := r.ensureUnitOwnedByAccount(ctx, query.OwnerAccountID, query.UnitID); err != nil {
		return nil, err
	}
	grant := r.q.UnitAccountRole
	q := grant.WithContext(ctx).
		Where(grant.UnitID.Eq(query.UnitID), grant.GrantedByAccountID.Eq(query.OwnerAccountID), grant.DeletedAt.IsNull()).
		Preload(grant.Role)
	if query.TargetAccountID > 0 {
		q = q.Where(grant.AccountID.Eq(query.TargetAccountID))
	}
	if query.Status != "" {
		q = q.Where(grant.Status.Eq(query.Status))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(grant.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainUnitAccountRoles(records), nil
}

func (r *RBACRepository) HasUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (bool, error) {
	unit := r.q.Unit
	unitRecord, err := unit.WithContext(ctx).Where(unit.ID.Eq(unitID), unit.DeletedAt.IsNull()).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, domainrbac.ErrUnitNotInAccount
		}
		return false, err
	}
	if unitRecord.AccountID == accountID {
		return true, nil
	}

	grant := r.q.UnitAccountRole
	role := r.q.Role
	roleGroup := r.q.RolePermissionGroup
	group := r.q.PermissionGroup
	groupPermission := r.q.PermissionGroupPermission
	permission := r.q.Permission
	count, err := grant.WithContext(ctx).
		Join(role, role.ID.EqCol(grant.RoleID)).
		Join(roleGroup, roleGroup.RoleID.EqCol(role.ID)).
		Join(group, group.ID.EqCol(roleGroup.PermissionGroupID)).
		Join(groupPermission, groupPermission.PermissionGroupID.EqCol(group.ID)).
		Join(permission, permission.ID.EqCol(groupPermission.PermissionID)).
		Where(
			grant.UnitID.Eq(unitID),
			grant.AccountID.Eq(accountID),
			grant.Status.Eq(domainrbac.UnitGrantStatusActive),
			grant.DeletedAt.IsNull(),
			role.IsActive.Is(true), role.DeletedAt.IsNull(),
			group.IsActive.Is(true), group.DeletedAt.IsNull(),
			permission.Code.Eq(permissionCode), permission.IsActive.Is(true), permission.DeletedAt.IsNull(),
		).Count()
	return count > 0, err
}

func (r *RBACRepository) ensureRoleInAccount(ctx context.Context, accountID int, roleID int) error {
	role := r.q.Role
	count, err := role.WithContext(ctx).Where(role.ID.Eq(roleID), role.AccountID.Eq(accountID), role.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainrbac.ErrRoleNotFound
	}
	return nil
}

func (r *RBACRepository) ensureRolesInAccount(ctx context.Context, accountID int, roleIDs []int) error {
	role := r.q.Role
	count, err := role.WithContext(ctx).Where(role.ID.In(roleIDs...), role.AccountID.Eq(accountID), role.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count != int64(len(roleIDs)) {
		return domainrbac.ErrRoleNotInAccount
	}
	return nil
}

func (r *RBACRepository) ensureUnitOwnedByAccount(ctx context.Context, accountID int, unitID int) error {
	unit := r.q.Unit
	count, err := unit.WithContext(ctx).Where(unit.ID.Eq(unitID), unit.AccountID.Eq(accountID), unit.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainrbac.ErrUnitNotInAccount
	}
	return nil
}

func (r *RBACRepository) ensureAccountExists(ctx context.Context, accountID int) error {
	account := r.q.Account
	count, err := account.WithContext(ctx).Where(account.ID.Eq(accountID), account.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainrbac.ErrAccountNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionGroupExists(ctx context.Context, groupID int) error {
	group := r.q.PermissionGroup
	count, err := group.WithContext(ctx).Where(group.ID.Eq(groupID), group.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainrbac.ErrPermissionGroupNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionGroupsExist(ctx context.Context, groupIDs []int) error {
	group := r.q.PermissionGroup
	count, err := group.WithContext(ctx).Where(group.ID.In(groupIDs...), group.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count != int64(len(groupIDs)) {
		return domainrbac.ErrPermissionGroupNotFound
	}
	return nil
}

func (r *RBACRepository) ensurePermissionsExist(ctx context.Context, permissionIDs []int) error {
	permission := r.q.Permission
	count, err := permission.WithContext(ctx).Where(permission.ID.In(permissionIDs...), permission.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count != int64(len(permissionIDs)) {
		return domainrbac.ErrPermissionNotFound
	}
	return nil
}

func toDomainRole(record *models.Role) *domainrbac.Role {
	if record == nil {
		return nil
	}
	return &domainrbac.Role{
		ID:             record.ID,
		AccountID:      record.AccountID,
		Name:           record.Name,
		IsDefaultOwner: record.IsDefaultOwner,
		IsActive:       *record.IsActive,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		DeletedAt:      record.DeletedAt,
	}
}

func toDomainRoles(records []*models.Role) []*domainrbac.Role {
	roles := make([]*domainrbac.Role, 0, len(records))
	for _, record := range records {
		roles = append(roles, toDomainRole(record))
	}
	return roles
}

func toDomainPermissionGroup(record *models.PermissionGroup) *domainrbac.PermissionGroup {
	if record == nil {
		return nil
	}
	return &domainrbac.PermissionGroup{
		ID:          record.ID,
		Code:        record.Code,
		Name:        record.Name,
		Description: record.Description,
		SortOrder:   record.SortOrder,
		IsActive:    *record.IsActive,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		DeletedAt:   record.DeletedAt,
	}
}

func toDomainPermissionGroups(records []*models.PermissionGroup) []*domainrbac.PermissionGroup {
	groups := make([]*domainrbac.PermissionGroup, 0, len(records))
	for _, record := range records {
		groups = append(groups, toDomainPermissionGroup(record))
	}
	return groups
}

func toDomainPermission(record *models.Permission) *domainrbac.Permission {
	if record == nil {
		return nil
	}
	return &domainrbac.Permission{
		ID:          record.ID,
		Code:        record.Code,
		Description: record.Description,
		IsActive:    *record.IsActive,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		DeletedAt:   record.DeletedAt,
	}
}

func toDomainPermissions(records []*models.Permission) []*domainrbac.Permission {
	permissions := make([]*domainrbac.Permission, 0, len(records))
	for _, record := range records {
		permissions = append(permissions, toDomainPermission(record))
	}
	return permissions
}

func toDomainUnitAccountRole(record *models.UnitAccountRole) *domainrbac.UnitAccountRole {
	if record == nil {
		return nil
	}
	grant := &domainrbac.UnitAccountRole{
		ID:                 record.ID,
		UnitID:             record.UnitID,
		AccountID:          record.AccountID,
		RoleID:             record.RoleID,
		GrantedByAccountID: record.GrantedByAccountID,
		Status:             record.Status,
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
		DeletedAt:          record.DeletedAt,
	}
	if record.Role != nil {
		grant.Role = toDomainRole(record.Role)
	}
	return grant
}

func toDomainUnitAccountRoles(records []*models.UnitAccountRole) []*domainrbac.UnitAccountRole {
	grants := make([]*domainrbac.UnitAccountRole, 0, len(records))
	for _, record := range records {
		grants = append(grants, toDomainUnitAccountRole(record))
	}
	return grants
}
