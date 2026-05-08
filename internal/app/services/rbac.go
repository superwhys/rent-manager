package services

import (
	"context"

	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	"github.com/superwhys/rent-manager-server/internal/infra/repo"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type RBACService struct {
	repo *repo.RBACRepository
}

func NewRBACService(repo *repo.RBACRepository) *RBACService {
	return &RBACService{repo: repo}
}

func (s *RBACService) CreateRoles(ctx context.Context, cmd *requests.CreateRolesCommand) ([]*entclient.Role, error) {
	params := make([]repo.CreateRoleParams, 0, len(cmd.Roles))
	for _, role := range cmd.Roles {
		params = append(params, repo.CreateRoleParams{
			Name:     role.Name,
			IsActive: true,
		})
	}
	return s.repo.CreateRoles(ctx, params)
}

func (s *RBACService) CreatePermissions(ctx context.Context, cmd *requests.CreatePermissionsCommand) ([]*entclient.Permission, error) {
	params := make([]repo.CreatePermissionParams, 0, len(cmd.Permissions))
	for _, permission := range cmd.Permissions {
		params = append(params, repo.CreatePermissionParams{
			Code:        permission.Code,
			APIPath:     permission.APIPath,
			Description: permission.Description,
			IsActive:    true,
		})
	}
	return s.repo.CreatePermissions(ctx, params)
}

func (s *RBACService) AddPermissionsToRole(ctx context.Context, cmd *requests.RolePermissionsCommand) error {
	return s.repo.AddPermissionsToRole(ctx, cmd.RoleID, cmd.PermissionIDs)
}

func (s *RBACService) RemovePermissionsFromRole(ctx context.Context, cmd *requests.RolePermissionsCommand) error {
	return s.repo.RemovePermissionsFromRole(ctx, cmd.RoleID, cmd.PermissionIDs)
}

func (s *RBACService) AssignRolesToAccount(ctx context.Context, cmd *requests.AccountRolesCommand) error {
	return s.repo.AssignRolesToAccount(ctx, cmd.AccountID, cmd.RoleIDs)
}

func (s *RBACService) HasPermission(ctx context.Context, accountID int, apiPath string) (bool, error) {
	return s.repo.HasPermission(ctx, accountID, apiPath)
}

func (s *RBACService) HasAccountRoles(ctx context.Context) (bool, error) {
	return s.repo.HasAccountRoles(ctx)
}
