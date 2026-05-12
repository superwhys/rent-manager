package account

import "context"

type Repository interface {
	GetByWechatOpenID(ctx context.Context, wechatOpenID string) (*Account, error)
	Create(ctx context.Context, account *Account) (*Account, error)
	Update(ctx context.Context, account *Account) (*Account, error)
	Delete(ctx context.Context, id int) error
}

type OrganizationRepository interface {
	CreateOrganization(ctx context.Context, organization *Organization, ownerAccountID int, ownerRoleName string, ownerPermissionGroupIDs []int) (*Organization, error)
	GetOrganization(ctx context.Context, organizationID int) (*Organization, error)
	ListOrganizationsByAccount(ctx context.Context, accountID int) ([]*Organization, error)
	ListAccounts(ctx context.Context, organizationID int) ([]*Account, error)
	AssignAccountToOrganization(ctx context.Context, organizationID int, accountID int) error
	UpdateAccountRoles(ctx context.Context, params AssignAccountRolesParams) error
	RemoveAccountFromOrganization(ctx context.Context, organizationID int, accountID int) error
}
