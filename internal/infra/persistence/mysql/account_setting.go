package mysql

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"

	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
)

// AccountSettingRepository 提供账号配置的 GORM 持久化实现。
type AccountSettingRepository struct {
	q *query.Query
}

// NewAccountSettingRepository 创建账号配置仓储。
func NewAccountSettingRepository(client *gorm.DB) *AccountSettingRepository {
	return &AccountSettingRepository{q: query.Use(client)}
}

// GetByKey 获取当前账号指定配置。
func (r *AccountSettingRepository) GetByKey(ctx context.Context, accountID string, key string) (*domainaccountsetting.AccountSetting, error) {
	a := r.q.AccountSetting
	record, err := a.WithContext(ctx).
		Where(a.AccountID.Eq(accountID), a.Key.Eq(key), a.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainaccountsetting.ErrNotFound
		}
		return nil, err
	}
	return toDomainAccountSetting(record), nil
}

// List 查询当前账号配置列表。
func (r *AccountSettingRepository) List(ctx context.Context, query domainaccountsetting.ListQuery) ([]*domainaccountsetting.AccountSetting, error) {
	a := r.q.AccountSetting
	records, err := a.WithContext(ctx).
		Where(a.AccountID.Eq(query.AccountID), a.DeletedAt.IsNull()).
		Order(a.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainAccountSettings(records), nil
}

// Upsert 创建或更新当前账号配置。
func (r *AccountSettingRepository) Upsert(ctx context.Context, setting *domainaccountsetting.AccountSetting) (*domainaccountsetting.AccountSetting, error) {
	if err := validateByteLength(64, setting.Key); err != nil {
		return nil, err
	}
	if err := validateByteLength(255, setting.Value); err != nil {
		return nil, err
	}
	a := r.q.AccountSetting
	record, err := a.WithContext(ctx).
		Where(a.AccountID.Eq(setting.AccountID), a.Key.Eq(setting.Key)).Take()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		record = &models.AccountSetting{AccountID: setting.AccountID, Key: setting.Key, Value: setting.Value}
		if err := a.WithContext(ctx).Create(record); err != nil {
			return nil, err
		}
		return toDomainAccountSetting(record), nil
	}
	if _, err := a.WithContext(ctx).Where(a.ID.Eq(record.ID)).
		UpdateSimple(a.Value.Value(setting.Value), a.DeletedAt.Null()); err != nil {
		return nil, err
	}
	return r.GetByKey(ctx, setting.AccountID, setting.Key)
}

func toDomainAccountSetting(record *models.AccountSetting) *domainaccountsetting.AccountSetting {
	if record == nil {
		return nil
	}
	return &domainaccountsetting.AccountSetting{
		ID:        record.ID,
		AccountID: record.AccountID,
		Key:       record.Key,
		Value:     record.Value,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}

func toDomainAccountSettings(records []*models.AccountSetting) []*domainaccountsetting.AccountSetting {
	settings := make([]*domainaccountsetting.AccountSetting, 0, len(records))
	for _, record := range records {
		settings = append(settings, toDomainAccountSetting(record))
	}
	return settings
}
