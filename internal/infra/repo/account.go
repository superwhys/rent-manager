package repo

import (
	"context"

	"github.com/superwhys/rent-manager-server/internal/domain/account"
	"github.com/superwhys/rent-manager-server/internal/ent"

	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
)

type AccountRepository struct {
	client *entclient.Client
}

func NewAccountRepository(client *entclient.Client) *AccountRepository {
	return &AccountRepository{
		client: client,
	}
}

func (r *AccountRepository) GetByWechatOpenID(ctx context.Context, wechatOpenID string) (*account.Account, error) {
	accountModel, err := r.client.Account.Query().
		Where(entaccount.WechatOpenid(wechatOpenID)).
		First(ctx)

	if ent.MaskNotFound(err) != nil {
		return nil, err
	}

	return toDomainAccount(accountModel), nil
}

func (r *AccountRepository) Create(ctx context.Context, createAccount *account.Account) (*account.Account, error) {
	accountModel, err := r.client.Account.Create().
		SetName(createAccount.Name).
		SetEmail(createAccount.Email).
		SetPassword(createAccount.Password).
		SetWechatOpenid(createAccount.WechatOpenID).
		SetIsActive(createAccount.IsActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toDomainAccount(accountModel), nil
}

func (r *AccountRepository) Update(ctx context.Context, updateAccount *account.Account) (*account.Account, error) {
	accountModel, err := r.client.Account.UpdateOneID(updateAccount.ID).
		SetName(updateAccount.Name).
		SetEmail(updateAccount.Email).
		SetPassword(updateAccount.Password).
		SetWechatOpenid(updateAccount.WechatOpenID).
		SetIsActive(updateAccount.IsActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toDomainAccount(accountModel), nil
}

func toDomainAccount(accountModel *entclient.Account) *account.Account {
	return &account.Account{
		ID:           accountModel.ID,
		Name:         accountModel.Name,
		Email:        accountModel.Email,
		WechatOpenID: accountModel.WechatOpenid,
		Password:     accountModel.Password,
		IsActive:     accountModel.IsActive,
		CreatedAt:    accountModel.CreatedAt,
		UpdatedAt:    accountModel.UpdatedAt,
		DeletedAt:    accountModel.DeletedAt,
	}
}
