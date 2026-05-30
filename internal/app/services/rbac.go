package services

import (
	"context"

	"git.superwhys.top/rent-manager/rent-manager-server/config"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"

	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
)

type RBACService struct {
	repo        domainrbac.Repository
	adminConfig *config.AdminConfig
}

func NewRBACService(repo domainrbac.Repository, adminConfig *config.AdminConfig) *RBACService {
	if adminConfig == nil {
		adminConfig = &config.AdminConfig{}
	}
	return &RBACService{repo: repo, adminConfig: adminConfig}
}

func (s *RBACService) CreateRoles(ctx context.Context, cmd *requests.CreateRolesCommand) ([]*domainrbac.Role, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roles := make([]*domainrbac.Role, 0, len(cmd.Roles))
	for _, item := range cmd.Roles {
		role, err := domainrbac.NewRole(domainrbac.CreateRoleParams{
			AccountID:      accountID,
			Name:           item.Name,
			IsDefaultOwner: item.IsDefaultOwner,
		})
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	created, err := s.repo.CreateRoles(ctx, roles)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *RBACService) ListRoles(ctx context.Context) ([]*domainrbac.Role, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.ListRoles(ctx, domainrbac.ListRolesQuery{AccountID: accountID})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *RBACService) GetRole(ctx context.Context, roleID int) (*domainrbac.Role, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.GetRole(ctx, accountID, roleID)
}

func (s *RBACService) UpdateRole(ctx context.Context, cmd *requests.UpdateRoleCommand) (*domainrbac.Role, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	role, err := s.repo.GetRole(ctx, accountID, cmd.RoleID)
	if err != nil {
		return nil, err
	}

	isActive := role.IsActive
	if cmd.IsActive != nil {
		isActive = *cmd.IsActive
	}

	if err := role.Update(domainrbac.UpdateRoleParams{
		ID:             cmd.RoleID,
		AccountID:      accountID,
		Name:           cmd.Name,
		IsDefaultOwner: cmd.IsDefaultOwner,
		IsActive:       isActive,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateRole(ctx, role)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *RBACService) DeleteRole(ctx context.Context, roleID int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteRole(ctx, accountID, roleID)
}

func (s *RBACService) AddPermissionGroupsToRole(ctx context.Context, cmd *requests.RolePermissionGroupsCommand) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	groupIDs := normalizeIDs(cmd.PermissionGroupIDs)
	if len(groupIDs) == 0 {
		return domainrbac.ErrPermissionGroupIDsRequired
	}

	return s.repo.AddPermissionGroupsToRole(ctx, accountID, cmd.RoleID, groupIDs)
}

func (s *RBACService) RemovePermissionGroupsFromRole(ctx context.Context, cmd *requests.RolePermissionGroupsCommand) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	groupIDs := normalizeIDs(cmd.PermissionGroupIDs)
	if len(groupIDs) == 0 {
		return domainrbac.ErrPermissionGroupIDsRequired
	}

	return s.repo.RemovePermissionGroupsFromRole(ctx, accountID, cmd.RoleID, groupIDs)
}

func (s *RBACService) CreatePermissionGroups(ctx context.Context, cmd *requests.CreatePermissionGroupsCommand) ([]*domainrbac.PermissionGroup, error) {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return nil, err
	}

	groups := make([]*domainrbac.PermissionGroup, 0, len(cmd.PermissionGroups))
	for _, item := range cmd.PermissionGroups {
		group, err := domainrbac.NewPermissionGroup(domainrbac.CreatePermissionGroupParams{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			SortOrder:   item.SortOrder,
		})
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	created, err := s.repo.CreatePermissionGroups(ctx, groups)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *RBACService) ListPermissionGroups(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	groups, err := s.repo.ListPermissionGroups(ctx)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *RBACService) CreatePermissions(ctx context.Context, cmd *requests.CreatePermissionsCommand) ([]*domainrbac.Permission, error) {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return nil, err
	}

	permissions := make([]*domainrbac.Permission, 0, len(cmd.Permissions))
	for _, item := range cmd.Permissions {
		permission, err := domainrbac.NewPermission(domainrbac.CreatePermissionParams{
			Code:        item.Code,
			Description: item.Description,
		})
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	created, err := s.repo.CreatePermissions(ctx, permissions)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]*domainrbac.Permission, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (s *RBACService) AddPermissionsToPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return err
	}

	permissionIDs := normalizeIDs(cmd.PermissionIDs)
	if len(permissionIDs) == 0 {
		return domainrbac.ErrPermissionNotFound
	}

	return s.repo.AddPermissionsToPermissionGroup(ctx, cmd.PermissionGroupID, permissionIDs)
}

func (s *RBACService) RemovePermissionsFromPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return err
	}

	permissionIDs := normalizeIDs(cmd.PermissionIDs)
	if len(permissionIDs) == 0 {
		return domainrbac.ErrPermissionNotFound
	}

	return s.repo.RemovePermissionsFromPermissionGroup(ctx, cmd.PermissionGroupID, permissionIDs)
}

func (s *RBACService) GrantUnitRoles(ctx context.Context, cmd *requests.UnitAccountRolesCommand) ([]*domainrbac.UnitAccountRole, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roleIDs := normalizeIDs(cmd.RoleIDs)
	if len(roleIDs) == 0 {
		return nil, domainrbac.ErrRoleIDsRequired
	}

	grants, err := s.repo.GrantUnitRoles(ctx, domainrbac.UnitAccountRolesParams{
		OwnerAccountID:  accountID,
		UnitID:          cmd.UnitID,
		TargetAccountID: cmd.AccountID,
		RoleIDs:         roleIDs,
	})
	if err != nil {
		return nil, err
	}

	return grants, nil
}

func (s *RBACService) RevokeUnitRoles(ctx context.Context, cmd *requests.UnitAccountRolesCommand) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	roleIDs := normalizeIDs(cmd.RoleIDs)
	if len(roleIDs) == 0 {
		return domainrbac.ErrRoleIDsRequired
	}

	return s.repo.RevokeUnitRoles(ctx, domainrbac.UnitAccountRolesParams{
		OwnerAccountID:  accountID,
		UnitID:          cmd.UnitID,
		TargetAccountID: cmd.AccountID,
		RoleIDs:         roleIDs,
	})
}

func (s *RBACService) ListUnitGrants(ctx context.Context, query *requests.UnitGrantsQuery) ([]*domainrbac.UnitAccountRole, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	grants, err := s.repo.ListUnitGrants(ctx, domainrbac.ListUnitGrantsQuery{
		OwnerAccountID:  accountID,
		UnitID:          query.UnitID,
		TargetAccountID: query.AccountID,
		Status:          query.Status,
		Limit:           query.Limit,
		Offset:          query.Offset,
	})
	if err != nil {
		return nil, err
	}

	return grants, nil
}

func (s *RBACService) RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error {
	if permissionCode == "" {
		return domainrbac.ErrPermissionCodeEmpty
	}

	allowed, err := s.repo.HasUnitPermission(ctx, accountID, unitID, permissionCode)
	if err != nil {
		return err
	}

	if !allowed {
		return domainrbac.ErrForbidden
	}

	return nil
}

func (s *RBACService) requireSuperAdmin(ctx context.Context) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	if !s.adminConfig.IsSuperAdmin(accountID) {
		return domainrbac.ErrForbidden
	}

	return nil
}

func normalizeIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}
