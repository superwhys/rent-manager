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
		role, err := domainrbac.NewRole(domainrbac.CreateRoleParams{Name: item.Name})
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return s.repo.CreateRoles(ctx, roles)
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

func (s *RBACService) AddPermissionsToRole(ctx context.Context, cmd *requests.RolePermissionsCommand) error {
	if err := s.repo.AddPermissionsToRole(ctx, cmd.RoleID, cmd.PermissionIDs); err != nil {
		return err
	}

	// 获取该角色分配到的所有账号，更新他们的缓存版本号
	accountIDs, err := s.repo.AccountIDsByRole(ctx, cmd.RoleID)
	if err != nil {
		return err
	}
	return s.updatePermissionCacheVersion(ctx, accountIDs...)
}

func (s *RBACService) RemovePermissionsFromRole(ctx context.Context, cmd *requests.RolePermissionsCommand) error {
	if err := s.repo.RemovePermissionsFromRole(ctx, cmd.RoleID, cmd.PermissionIDs); err != nil {
		return err
	}

	// 获取该角色分配到的所有账号，更新他们的缓存版本号
	accountIDs, err := s.repo.AccountIDsByRole(ctx, cmd.RoleID)
	if err != nil {
		return err
	}
	return s.updatePermissionCacheVersion(ctx, accountIDs...)
}

func (s *RBACService) AssignRolesToAccount(ctx context.Context, cmd *requests.AccountRolesCommand) error {
	if err := s.repo.AssignRolesToAccount(ctx, cmd.AccountID, cmd.RoleIDs); err != nil {
		return err
	}

	return s.updatePermissionCacheVersion(ctx, cmd.AccountID)
}

// HasPermission 检查一个 account 在 apiPath 接口上有没有权限
func (s *RBACService) HasPermission(ctx context.Context, accountID int, apiPath string) (bool, error) {
	version, cacheOK := s.permissionCacheVersion(ctx, accountID)
	req := requests.PermissionCheckRequest{
		AccountID: accountID,
		APIPath:   apiPath,
		Version:   version,
	}
	if !cacheOK {
		return s.hasPermissionCheck(ctx, req)
	}

	return s.hasPermissionCheckCachedFn(ctx, req)
}

func (s *RBACService) hasPermissionCheck(ctx context.Context, req requests.PermissionCheckRequest) (bool, error) {
	return s.repo.HasPermission(ctx, req.AccountID, req.APIPath)
}

// updatePermissionCacheVersion 更新 accountIDs 账号权限的版本号缓存
func (s *RBACService) updatePermissionCacheVersion(ctx context.Context, accountIDs ...int) error {
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
		pipe.Incr(ctx, cache.PermissionVersionCache(strconv.Itoa(accountID)).Key())
	}
	_, err := pipe.Exec(ctx)
	return err
}

// permissionCacheVersion 获取一个账号的权限版本号
func (s *RBACService) permissionCacheVersion(ctx context.Context, accountID int) (int64, bool) {
	if s.redisClient == nil {
		return 0, false
	}

	version, err := cache.PermissionVersionCache(strconv.Itoa(accountID)).Get(ctx, s.redisClient)
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
