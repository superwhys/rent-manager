package services

import (
	"context"

	domainaccount "github.com/superwhys/rent-manager-server/internal/domain/account"
	"github.com/superwhys/rent-manager-server/internal/models/consts"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

// OrganizationService 编排组织创建和组织账号绑定用例。
type OrganizationService struct {
	repo        domainaccount.OrganizationRepository
	rbacService *RBACService
}

func NewOrganizationService(repo domainaccount.OrganizationRepository, rbacService *RBACService) *OrganizationService {
	return &OrganizationService{repo: repo, rbacService: rbacService}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, cmd *requests.CreateOrganizationCommand) (*domainaccount.Organization, error) {
	organization, err := domainaccount.NewOrganization(domainaccount.CreateOrganizationParams{Name: cmd.Name})
	if err != nil {
		return nil, err
	}

	organization, err = s.repo.CreateOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	roles, err := s.rbacService.CreateRoles(ctx, &requests.CreateRolesCommand{
		OrganizationID: organization.ID,
		Roles: []requests.CreateRoleItem{
			{
				Name:           consts.OrganizationRoleAdmin,
				IsDefaultOwner: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(cmd.AdminPermissionGroupIDs) > 0 && len(roles) > 0 {
		err = s.rbacService.AddPermissionGroupsToRole(ctx, &requests.RolePermissionGroupsCommand{
			OrganizationID:     organization.ID,
			RoleID:             roles[0].ID,
			PermissionGroupIDs: cmd.AdminPermissionGroupIDs,
		})
		if err != nil {
			return nil, err
		}
	}
	return organization, nil
}

func (s *OrganizationService) BindAccountToOrganization(ctx context.Context, cmd *requests.BindAccountToOrganizationCommand) error {
	return s.repo.AssignAccountToOrganization(ctx, cmd.OrganizationID, cmd.AccountID)
}
