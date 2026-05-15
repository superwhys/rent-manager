package services

import (
	"context"

	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/rent-manager-server/internal/models/requests"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
)

type RBACService struct {
	repo        domainrbac.Repository
	redisClient *redisutils.RedisClient
}

func NewRBACService(repo domainrbac.Repository, redisClient *redisutils.RedisClient) *RBACService {
	return &RBACService{repo: repo, redisClient: redisClient}
}

func (s *RBACService) InitSystemRBAC(ctx context.Context) error {
	return s.repo.InitSystemRBAC(ctx)
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
	return s.repo.CreateRoles(ctx, roles)
}

func (s *RBACService) CreatePermissionGroups(ctx context.Context, cmd *requests.CreatePermissionGroupsCommand) ([]*domainrbac.PermissionGroup, error) {
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
	return s.repo.CreatePermissionGroups(ctx, groups)
}

func (s *RBACService) CreatePermissions(ctx context.Context, cmd *requests.CreatePermissionsCommand) ([]*domainrbac.Permission, error) {
	permissions := make([]*domainrbac.Permission, 0, len(cmd.Permissions))
	for _, item := range cmd.Permissions {
		permission, err := domainrbac.NewPermission(domainrbac.CreatePermissionParams{
			Code:        item.Code,
			APIPath:     item.APIPath,
			Description: item.Description,
		})
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return s.repo.CreatePermissions(ctx, permissions)
}

func (s *RBACService) AddPermissionsToPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	return s.repo.AddPermissionsToPermissionGroup(ctx, cmd.PermissionGroupID, cmd.PermissionIDs)
}

func (s *RBACService) RemovePermissionsFromPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	return s.repo.RemovePermissionsFromPermissionGroup(ctx, cmd.PermissionGroupID, cmd.PermissionIDs)
}

func (s *RBACService) AddPermissionGroupsToRole(ctx context.Context, cmd *requests.RolePermissionGroupsCommand) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.AddPermissionGroupsToRole(ctx, accountID, cmd.RoleID, cmd.PermissionGroupIDs)
}

func (s *RBACService) RemovePermissionGroupsFromRole(ctx context.Context, cmd *requests.RolePermissionGroupsCommand) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.RemovePermissionGroupsFromRole(ctx, accountID, cmd.RoleID, cmd.PermissionGroupIDs)
}

func (s *RBACService) AssignRolesToUnitAccount(ctx context.Context, cmd *requests.UnitAccountRolesCommand) error {
	ownerAccountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.AssignRolesToUnitAccount(ctx, ownerAccountID, cmd.UnitID, cmd.AccountID, cmd.RoleIDs)
}

func (s *RBACService) HasSystemPermission(ctx context.Context, accountID int) (bool, error) {
	return s.repo.HasSystemPermission(ctx, accountID)
}
