package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/superwhys/rent-manager/internal/domain/account"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type AccountRepository struct {
	q *query.Query
}

func NewAccountRepository(client *gorm.DB) *AccountRepository {
	return &AccountRepository{q: query.Use(client)}
}

func (r *AccountRepository) Get(ctx context.Context, id int) (*account.Account, error) {
	a := r.q.Account
	record, err := a.WithContext(ctx).
		Where(a.ID.Eq(id), a.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) GetByWechatOpenID(ctx context.Context, wechatOpenID string) (*account.Account, error) {
	a := r.q.Account
	record, err := a.WithContext(ctx).
		Where(a.WechatOpenid.Eq(wechatOpenID), a.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Create(ctx context.Context, createAccount *account.Account) (*account.Account, error) {
	if err := validateByteLength(255, createAccount.Name, createAccount.Password, createAccount.WechatOpenID); err != nil {
		return nil, err
	}
	record := &models.Account{
		Name:         createAccount.Name,
		Email:        createAccount.Email,
		Password:     createAccount.Password,
		WechatOpenid: createAccount.WechatOpenID,
		IsActive:     &createAccount.IsActive,
	}
	if err := r.q.Account.WithContext(ctx).Create(record); err != nil {
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Update(ctx context.Context, updateAccount *account.Account) (*account.Account, error) {
	if err := validateByteLength(255, updateAccount.Name, updateAccount.Password, updateAccount.WechatOpenID); err != nil {
		return nil, err
	}
	a := r.q.Account
	_, err := a.WithContext(ctx).Where(a.ID.Eq(updateAccount.ID)).
		UpdateSimple(
			a.Name.Value(updateAccount.Name),
			a.Email.Value(updateAccount.Email),
			a.Password.Value(updateAccount.Password),
			a.WechatOpenid.Value(updateAccount.WechatOpenID),
			a.IsActive.Value(updateAccount.IsActive),
		)
	if err != nil {
		return nil, err
	}

	record, err := a.WithContext(ctx).Where(a.ID.Eq(updateAccount.ID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(record), nil
}

func (r *AccountRepository) Delete(ctx context.Context, id int) error {
	a := r.q.Account
	result, err := a.WithContext(ctx).
		Where(a.ID.Eq(id), a.DeletedAt.IsNull()).
		Update(a.DeletedAt, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return account.ErrAccountNotFound
	}
	return nil
}

func toDomainAccount(record *models.Account) *account.Account {
	if record == nil {
		return nil
	}
	return &account.Account{
		ID:           record.ID,
		Name:         record.Name,
		Email:        record.Email,
		Password:     record.Password,
		WechatOpenID: record.WechatOpenid,
		IsActive:     *record.IsActive,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		DeletedAt:    record.DeletedAt,
	}
}
