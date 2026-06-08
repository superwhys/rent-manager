package rbac

import "context"

var _ IRbacService = (*RbacService)(nil)

type IRbacService interface {
	RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error
}

type RbacService struct {
	repo Repository
}

func NewRbacService(repo Repository) *RbacService {
	return &RbacService{repo: repo}
}

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
