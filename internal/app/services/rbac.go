package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/miebyte/goutils/redisutils"
	"github.com/redis/go-redis/v9"
	"github.com/superwhys/rent-manager-server/internal/infra/cache"
	"github.com/superwhys/rent-manager-server/internal/models/requests"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
)

type RBACService struct {
	repo                       domainrbac.Repository
	redisClient                *redisutils.RedisClient
	hasPermissionCheckCachedFn redisutils.AnyFn[requests.PermissionCheckRequest, bool]
}

func NewRBACService(repo domainrbac.Repository, redisClient *redisutils.RedisClient) *RBACService {
	s := &RBACService{repo: repo, redisClient: redisClient}
	s.hasPermissionCheckCachedFn = redisutils.CacheCall(s.hasPermissionCheck).
		WithClient(redisClient).
		SetPrefix("rent-manager:rbac:permission").
		SetTTL(time.Minute * 5).
		SetTTLJitter(time.Minute).
		OnceFunc()
	return s
}

func (s *RBACService) InitSystemRBAC(ctx context.Context) error {
	return s.repo.InitSystemRBAC(ctx)
}

func (s *RBACService) CreateRoles(ctx context.Context, cmd *requests.CreateRolesCommand) ([]*domainrbac.Role, error) {
	roles := make([]*domainrbac.Role, 0, len(cmd.Roles))
	for _, item := range cmd.Roles {
		role, err := domainrbac.NewRole(domainrbac.CreateRoleParams{
			OrganizationID: cmd.OrganizationID,
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
	if err := s.repo.AddPermissionGroupsToRole(ctx, cmd.OrganizationID, cmd.RoleID, cmd.PermissionGroupIDs); err != nil {
		return err
	}

	accountIDs, err := s.repo.AccountIDsByRole(ctx, cmd.OrganizationID, cmd.RoleID)
	if err != nil {
		return err
	}
	return s.updatePermissionCacheVersion(ctx, cmd.OrganizationID, accountIDs...)
}

func (s *RBACService) RemovePermissionGroupsFromRole(ctx context.Context, cmd *requests.RolePermissionGroupsCommand) error {
	if err := s.repo.RemovePermissionGroupsFromRole(ctx, cmd.OrganizationID, cmd.RoleID, cmd.PermissionGroupIDs); err != nil {
		return err
	}

	accountIDs, err := s.repo.AccountIDsByRole(ctx, cmd.OrganizationID, cmd.RoleID)
	if err != nil {
		return err
	}
	return s.updatePermissionCacheVersion(ctx, cmd.OrganizationID, accountIDs...)
}

func (s *RBACService) AssignRolesToAccount(ctx context.Context, cmd *requests.AccountRolesCommand) error {
	if err := s.repo.AssignRolesToOrganizationAccount(ctx, cmd.OrganizationID, cmd.AccountID, cmd.RoleIDs); err != nil {
		return err
	}

	return s.updatePermissionCacheVersion(ctx, cmd.OrganizationID, cmd.AccountID)
}

func (s *RBACService) HasPermission(ctx context.Context, accountID int, organizationID int, apiPath string) (bool, error) {
	version, cacheOK := s.permissionCacheVersion(ctx, accountID, organizationID)
	req := requests.PermissionCheckRequest{
		AccountID:      accountID,
		OrganizationID: organizationID,
		APIPath:        apiPath,
		Version:        version,
	}
	if !cacheOK {
		return s.hasPermissionCheck(ctx, req)
	}

	return s.hasPermissionCheckCachedFn(ctx, req)
}

func (s *RBACService) hasPermissionCheck(ctx context.Context, req requests.PermissionCheckRequest) (bool, error) {
	return s.repo.HasPermission(ctx, req.AccountID, req.OrganizationID, req.APIPath)
}

func (s *RBACService) updatePermissionCacheVersion(ctx context.Context, organizationID int, accountIDs ...int) error {
	if s.redisClient == nil || len(accountIDs) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(accountIDs))
	pipe := s.redisClient.Pipeline()
	for _, accountID := range accountIDs {
		if _, ok := seen[accountID]; ok {
			continue
		}
		seen[accountID] = struct{}{}
		pipe.Incr(ctx, cache.PermissionVersionCache(permissionCacheVersionKey(accountID, organizationID)).Key())
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RBACService) permissionCacheVersion(ctx context.Context, accountID int, organizationID int) (int64, bool) {
	if s.redisClient == nil {
		return 0, false
	}

	version, err := cache.PermissionVersionCache(permissionCacheVersionKey(accountID, organizationID)).Get(ctx, s.redisClient)
	if err == nil {
		versionID, parseErr := strconv.ParseInt(version, 10, 64)
		if parseErr != nil {
			return 0, false
		}
		return versionID, true
	}

	if errors.Is(err, redis.Nil) {
		return 0, true
	}
	return 0, false
}

func permissionCacheVersionKey(accountID int, organizationID int) string {
	return strconv.Itoa(accountID) + ":" + strconv.Itoa(organizationID)
}
