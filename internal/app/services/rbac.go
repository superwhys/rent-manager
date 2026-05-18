package services

import (
	"context"
	"errors"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	"github.com/superwhys/rent-manager-server/internal/errcode"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type RBACService struct {
	repo domainrbac.Repository
}

func NewRBACService(repo domainrbac.Repository) *RBACService {
	return &RBACService{repo: repo}
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
			return nil, mapRBACError(err)
		}
		groups = append(groups, group)
	}

	created, err := s.repo.CreatePermissionGroups(ctx, groups)
	if err != nil {
		return nil, mapRBACError(err)
	}
	return created, nil
}

func (s *RBACService) ListPermissionGroups(ctx context.Context) ([]*domainrbac.PermissionGroup, error) {
	groups, err := s.repo.ListPermissionGroups(ctx)
	if err != nil {
		return nil, mapRBACError(err)
	}
	return groups, nil
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
			return nil, mapRBACError(err)
		}
		permissions = append(permissions, permission)
	}

	created, err := s.repo.CreatePermissions(ctx, permissions)
	if err != nil {
		return nil, mapRBACError(err)
	}
	return created, nil
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]*domainrbac.Permission, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, mapRBACError(err)
	}
	return permissions, nil
}

func (s *RBACService) AddPermissionsToPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	cmd.PermissionIDs = normalizePermissionIDs(cmd.PermissionIDs)
	return mapRBACError(s.repo.AddPermissionsToPermissionGroup(ctx, cmd.PermissionGroupID, cmd.PermissionIDs))
}

func (s *RBACService) RemovePermissionsFromPermissionGroup(ctx context.Context, cmd *requests.PermissionGroupPermissionsCommand) error {
	cmd.PermissionIDs = normalizePermissionIDs(cmd.PermissionIDs)
	return mapRBACError(s.repo.RemovePermissionsFromPermissionGroup(ctx, cmd.PermissionGroupID, cmd.PermissionIDs))
}

func normalizePermissionIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}

func mapRBACError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domainrbac.ErrPermissionGroupNotFound),
		errors.Is(err, domainrbac.ErrPermissionNotFound):
		return errcode.ErrNotFound.WithMessage(err.Error())
	case errors.Is(err, domainrbac.ErrPermissionGroupCodeRequired),
		errors.Is(err, domainrbac.ErrPermissionGroupNameRequired),
		errors.Is(err, domainrbac.ErrPermissionCodeRequired),
		errors.Is(err, domainrbac.ErrPermissionAPIPathRequired),
		errors.Is(err, domainrbac.ErrPermissionDescriptionRequired),
		errors.Is(err, domainrbac.ErrPermissionGroupAlreadyExists),
		errors.Is(err, domainrbac.ErrPermissionAlreadyExists):
		return errcode.ErrBadRequest.WithMessage(err.Error())
	default:
		return err
	}
}
