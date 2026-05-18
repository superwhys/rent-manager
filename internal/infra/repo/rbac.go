package repo

import (
	"context"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entpermission "github.com/superwhys/rent-manager-server/internal/ent/permission"
	entpermissiongroup "github.com/superwhys/rent-manager-server/internal/ent/permissiongroup"
)

type RBACRepository struct {
	client *entclient.Client
}

func NewRBACRepository(client *entclient.Client) *RBACRepository {
	return &RBACRepository{client: client}
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
			SetAPIPath(permission.APIPath).
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
