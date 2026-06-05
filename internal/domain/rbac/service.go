package rbac

import "context"

var _ IRbacService = (*RbacService)(nil)

type IRbacService interface {
	RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error
}

// RbacService 提供 RBAC 领域服务。
type RbacService struct {
	repo Repository
}

// NewRbacService 创建 RBAC 领域服务。
func NewRbacService(repo Repository) *RbacService {
	return &RbacService{repo: repo}
}

// RequireUnitPermission 校验账号是否拥有指定单元权限。
func (s *RbacService) RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error {
	if permissionCode == "" {
		return ErrPermissionCodeEmpty
	}

	allowed, err := s.repo.HasUnitPermission(ctx, accountID, unitID, permissionCode)
	if err != nil {
		return err
	}

	if !allowed {
		return ErrForbidden
	}

	return nil
}
