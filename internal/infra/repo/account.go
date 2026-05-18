package repo

import (
	"context"

	"github.com/superwhys/rent-manager-server/internal/domain/account"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entaccount "github.com/superwhys/rent-manager-server/internal/ent/account"
)

type AccountRepository struct {
	client *entclient.Client
}

func NewAccountRepository(client *entclient.Client) *AccountRepository {
	return &AccountRepository{client: client}
}

func (r *AccountRepository) GetByWechatOpenID(ctx context.Context, wechatOpenID string) (*account.Account, error) {
	record, err := r.client.Account.Query().
		Where(entaccount.WechatOpenid(wechatOpenID), entaccount.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Create(ctx context.Context, createAccount *account.Account) (*account.Account, error) {
	record, err := r.client.Account.Create().
		SetName(createAccount.Name).
		SetEmail(createAccount.Email).
		SetPassword(createAccount.Password).
		SetWechatOpenid(createAccount.WechatOpenID).
		SetIsActive(createAccount.IsActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Update(ctx context.Context, updateAccount *account.Account) (*account.Account, error) {
	record, err := r.client.Account.UpdateOneID(updateAccount.ID).
		SetName(updateAccount.Name).
		SetEmail(updateAccount.Email).
		SetPassword(updateAccount.Password).
		SetWechatOpenid(updateAccount.WechatOpenID).
		SetIsActive(updateAccount.IsActive).
		Save(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Delete(ctx context.Context, id int) error {
	affected, err := r.client.Account.Update().
		Where(entaccount.ID(id), entaccount.DeletedAtIsNil()).
		SetDeletedAt(entaccount.DefaultUpdatedAt()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return account.ErrAccountNotFound
	}
	return nil
}

func toDomainAccount(record *entclient.Account) *account.Account {
	if record == nil {
		return nil
	}
	return &account.Account{
		ID:           record.ID,
		Name:         record.Name,
		Email:        record.Email,
		Password:     record.Password,
		WechatOpenID: record.WechatOpenid,
		IsActive:     record.IsActive,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		DeletedAt:    record.DeletedAt,
	}
}
