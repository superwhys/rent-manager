package rbac

import (
	"context"
	"errors"

	"github.com/miebyte/authkit"

	"github.com/superwhys/rent-manager/config"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供角色、权限和单元授权用例。
type Service struct {
	repo        domainrbac.Repository
	accounts    AccountLookup
	units       UnitLookup
	adminConfig *config.AdminConfig
}

// AccountLookup 直接查询 authkit 的账号，不维护本地账号或 ID 映射。
type AccountLookup interface {
	GetByID(context.Context, string) (*authkit.Account, error)
}

type UnitLookup interface {
	Get(context.Context, string, int) (*domainunit.Unit, error)
}

// NewService 创建 RBAC 服务。
func NewService(repo domainrbac.Repository, accounts AccountLookup, units UnitLookup, adminConfig *config.AdminConfig) *Service {
	if adminConfig == nil {
		adminConfig = &config.AdminConfig{}
	}
	return &Service{
		repo:        repo,
		accounts:    accounts,
		units:       units,
		adminConfig: adminConfig,
	}
}

// CreateRoles 批量创建当前账号下的角色。
func (s *Service) CreateRoles(ctx context.Context, cmd *CreateRolesCommand) ([]*domainrbac.Role, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// ListRoles 获取当前账号角色列表。
func (s *Service) ListRoles(ctx context.Context) ([]*domainrbac.Role, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.ListRoles(ctx, domainrbac.ListRolesQuery{AccountID: accountID})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRole 获取当前账号角色详情。
func (s *Service) GetRole(ctx context.Context, roleID int) (*domainrbac.Role, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.GetRole(ctx, accountID, roleID)
}

// UpdateRole 更新当前账号下的角色。
func (s *Service) UpdateRole(ctx context.Context, cmd *UpdateRoleCommand) (*domainrbac.Role, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// DeleteRole 删除当前账号下的角色。
func (s *Service) DeleteRole(ctx context.Context, roleID int) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteRole(ctx, accountID, roleID)
}

// AddPermissionGroupsToRole 给当前账号角色绑定权限组。
func (s *Service) AddPermissionGroupsToRole(ctx context.Context, cmd *RolePermissionGroupsCommand) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	groupIDs := normalizeIDs(cmd.PermissionGroupIDs)
	if len(groupIDs) == 0 {
		return domainrbac.ErrPermissionGroupIDsRequired
	}

	return s.repo.AddPermissionGroupsToRole(ctx, accountID, cmd.RoleID, groupIDs)
}

// RemovePermissionGroupsFromRole 从当前账号角色移除权限组。
func (s *Service) RemovePermissionGroupsFromRole(ctx context.Context, cmd *RolePermissionGroupsCommand) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	groupIDs := normalizeIDs(cmd.PermissionGroupIDs)
	if len(groupIDs) == 0 {
		return domainrbac.ErrPermissionGroupIDsRequired
	}

	return s.repo.RemovePermissionGroupsFromRole(ctx, accountID, cmd.RoleID, groupIDs)
}

// CreatePermissionGroups 批量创建全局权限组。
func (s *Service) CreatePermissionGroups(ctx context.Context, cmd *CreatePermissionGroupsCommand) ([]*domainrbac.PermissionGroup, error) {
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

// ListPermissionGroups 获取全局权限组列表。
func (s *Service) ListPermissionGroups(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	groups, err := s.repo.ListPermissionGroups(ctx)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

// CreatePermissions 批量创建全局权限。
func (s *Service) CreatePermissions(ctx context.Context, cmd *CreatePermissionsCommand) ([]*domainrbac.Permission, error) {
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

// ListPermissions 获取全局权限列表。
func (s *Service) ListPermissions(ctx context.Context) ([]*domainrbac.Permission, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddPermissionsToPermissionGroup 给权限组绑定权限。
func (s *Service) AddPermissionsToPermissionGroup(ctx context.Context, cmd *PermissionGroupPermissionsCommand) error {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return err
	}

	permissionIDs := normalizeIDs(cmd.PermissionIDs)
	if len(permissionIDs) == 0 {
		return domainrbac.ErrPermissionNotFound
	}

	return s.repo.AddPermissionsToPermissionGroup(ctx, cmd.PermissionGroupID, permissionIDs)
}

// RemovePermissionsFromPermissionGroup 从权限组移除权限。
func (s *Service) RemovePermissionsFromPermissionGroup(ctx context.Context, cmd *PermissionGroupPermissionsCommand) error {
	if err := s.requireSuperAdmin(ctx); err != nil {
		return err
	}

	permissionIDs := normalizeIDs(cmd.PermissionIDs)
	if len(permissionIDs) == 0 {
		return domainrbac.ErrPermissionNotFound
	}

	return s.repo.RemovePermissionsFromPermissionGroup(ctx, cmd.PermissionGroupID, permissionIDs)
}

// GrantUnitRoles 给目标账号授予单元角色。
func (s *Service) GrantUnitRoles(ctx context.Context, cmd *UnitAccountRolesCommand) ([]*domainrbac.UnitAccountRole, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	roleIDs := normalizeIDs(cmd.RoleIDs)
	if len(roleIDs) == 0 {
		return nil, domainrbac.ErrRoleIDsRequired
	}
	if cmd.AccountID == "" {
		return nil, domainrbac.ErrTargetAccountRequired
	}
	// 先确认单元归属，避免向无权操作的调用者泄露目标账号是否存在。
	if _, err := s.units.Get(ctx, accountID, cmd.UnitID); err != nil {
		if errors.Is(err, domainunit.ErrNotFound) {
			return nil, domainrbac.ErrUnitNotInAccount
		}
		return nil, err
	}
	if _, err := s.accounts.GetByID(ctx, cmd.AccountID); err != nil {
		if errors.Is(err, authkit.ErrNotFound) {
			return nil, domainrbac.ErrAccountNotFound
		}
		return nil, err
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

// RevokeUnitRoles 撤销目标账号的单元角色。
func (s *Service) RevokeUnitRoles(ctx context.Context, cmd *UnitAccountRolesCommand) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// ListUnitGrants 获取当前账号作为所有者的单元授权列表。
func (s *Service) ListUnitGrants(ctx context.Context, query *UnitGrantsQuery) ([]*domainrbac.UnitAccountRole, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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

// requireSuperAdmin 校验当前账号是否为超级管理员。
func (s *Service) requireSuperAdmin(ctx context.Context) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	if !s.adminConfig.IsSuperAdmin(accountID) {
		return domainrbac.ErrForbidden
	}

	return nil
}

// normalizeIDs 过滤非正数并去重 ID 列表。
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
