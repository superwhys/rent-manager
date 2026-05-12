package repo

import (
	"context"

	domainaccount "github.com/superwhys/rent-manager-server/internal/domain/account"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
	entorganization "github.com/superwhys/rent-manager-server/internal/ent/organization"
	entrole "github.com/superwhys/rent-manager-server/internal/ent/role"
)

type OrganizationRepository struct {
	client *entclient.Client
}

func NewOrganizationRepository(client *entclient.Client) *OrganizationRepository {
	return &OrganizationRepository{client: client}
}

func (r *OrganizationRepository) CreateOrganization(ctx context.Context, organization *domainaccount.Organization, ownerAccountID int, ownerRoleName string, ownerPermissionGroupIDs []int) (*domainaccount.Organization, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	organizationRecord, err := tx.Organization.Create().
		SetName(organization.Name).
		SetIsActive(organization.IsActive).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, mapOrganizationError(err)
	}

	roleBuilder := tx.Role.Create().
		SetOrganizationID(organizationRecord.ID).
		SetName(ownerRoleName).
		SetIsDefaultOwner(true).
		SetIsActive(true)
	if len(ownerPermissionGroupIDs) > 0 {
		roleBuilder = roleBuilder.AddPermissionGroupIDs(ownerPermissionGroupIDs...)
	}
	roleRecord, err := roleBuilder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	_, err = tx.Account.UpdateOneID(ownerAccountID).
		SetOrganizationID(organizationRecord.ID).
		AddRoleIDs(roleRecord.ID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return toDomainOrganization(organizationRecord), nil
}

func (r *OrganizationRepository) GetOrganization(ctx context.Context, organizationID int) (*domainaccount.Organization, error) {
	record, err := r.client.Organization.Query().
		Where(entorganization.ID(organizationID), entorganization.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, mapOrganizationError(err)
	}
	return toDomainOrganization(record), nil
}

func (r *OrganizationRepository) ListOrganizationsByAccount(ctx context.Context, accountID int) ([]*domainaccount.Organization, error) {
	records, err := r.client.Organization.Query().
		Where(
			entorganization.DeletedAtIsNil(),
			entorganization.HasAccountsWith(entaccount.ID(accountID), entaccount.DeletedAtIsNil()),
		).
		Order(entclient.Desc(entorganization.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	organizations := make([]*domainaccount.Organization, 0, len(records))
	for _, record := range records {
		organizations = append(organizations, toDomainOrganization(record))
	}
	return organizations, nil
}

func (r *OrganizationRepository) ListAccounts(ctx context.Context, organizationID int) ([]*domainaccount.Account, error) {
	records, err := r.client.Account.Query().
		Where(entaccount.OrganizationID(organizationID), entaccount.DeletedAtIsNil()).
		Order(entclient.Desc(entaccount.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	accounts := make([]*domainaccount.Account, 0, len(records))
	for _, record := range records {
		accounts = append(accounts, toDomainAccount(record))
	}
	return accounts, nil
}

func (r *OrganizationRepository) AssignAccountToOrganization(ctx context.Context, organizationID int, accountID int) error {
	return r.client.Account.UpdateOneID(accountID).
		SetOrganizationID(organizationID).
		Exec(ctx)
}

func (r *OrganizationRepository) UpdateAccountRoles(ctx context.Context, params domainaccount.AssignAccountRolesParams) error {
	count, err := r.client.Role.Query().
		Where(entrole.IDIn(params.RoleIDs...), entrole.OrganizationID(params.OrganizationID), entrole.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(params.RoleIDs) {
		return domainaccount.ErrAccountNotInOrganization
	}

	affected, err := r.client.Account.Update().
		Where(entaccount.ID(params.AccountID), entaccount.OrganizationID(params.OrganizationID), entaccount.DeletedAtIsNil()).
		ClearRoles().
		AddRoleIDs(params.RoleIDs...).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainaccount.ErrAccountNotInOrganization
	}
	return nil
}

func (r *OrganizationRepository) RemoveAccountFromOrganization(ctx context.Context, organizationID int, accountID int) error {
	affected, err := r.client.Account.Update().
		Where(entaccount.ID(accountID), entaccount.OrganizationID(organizationID), entaccount.DeletedAtIsNil()).
		ClearOrganization().
		ClearRoles().
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainaccount.ErrAccountNotInOrganization
	}
	return nil
}

func toDomainOrganization(record *entclient.Organization) *domainaccount.Organization {
	if record == nil {
		return nil
	}
	return &domainaccount.Organization{
		ID:        record.ID,
		Name:      record.Name,
		IsActive:  record.IsActive,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}

func mapOrganizationError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domainaccount.ErrOrganizationNotFound
	}
	return err
}
