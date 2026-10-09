package mysql

import (
	"context"

	domainaccountsetting "github.com/superwhys/rent-manager/internal/domain/accountsetting"
	entclient "github.com/superwhys/rent-manager/internal/infra/persistence/ent"
	entaccountsetting "github.com/superwhys/rent-manager/internal/infra/persistence/ent/accountsetting"
)

// AccountSettingRepository 提供账号配置的 Ent 持久化实现。
type AccountSettingRepository struct {
	client *entclient.Client
}

// NewAccountSettingRepository 创建账号配置仓储。
func NewAccountSettingRepository(client *entclient.Client) *AccountSettingRepository {
	return &AccountSettingRepository{client: client}
}

// GetByKey 获取当前账号指定配置。
func (r *AccountSettingRepository) GetByKey(ctx context.Context, accountID int, key string) (*domainaccountsetting.AccountSetting, error) {
	record, err := r.client.AccountSetting.Query().
		Where(
			entaccountsetting.AccountID(accountID),
			entaccountsetting.Key(key),
			entaccountsetting.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainaccountsetting.ErrNotFound
		}
		return nil, err
	}
	return toDomainAccountSetting(record), nil
}

// List 查询当前账号配置列表。
func (r *AccountSettingRepository) List(ctx context.Context, query domainaccountsetting.ListQuery) ([]*domainaccountsetting.AccountSetting, error) {
	records, err := r.client.AccountSetting.Query().
		Where(
			entaccountsetting.AccountID(query.AccountID),
			entaccountsetting.DeletedAtIsNil(),
		).
		Order(entclient.Desc(entaccountsetting.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainAccountSettings(records), nil
}

// Upsert 创建或更新当前账号配置。
func (r *AccountSettingRepository) Upsert(ctx context.Context, setting *domainaccountsetting.AccountSetting) (*domainaccountsetting.AccountSetting, error) {
	record, err := r.client.AccountSetting.Query().
		Where(
			entaccountsetting.AccountID(setting.AccountID),
			entaccountsetting.Key(setting.Key),
		).
		Only(ctx)
	if err != nil {
		if !entclient.IsNotFound(err) {
			return nil, err
		}
		created, err := r.client.AccountSetting.Create().
			SetAccountID(setting.AccountID).
			SetKey(setting.Key).
			SetValue(setting.Value).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		return toDomainAccountSetting(created), nil
	}

	updated, err := r.client.AccountSetting.UpdateOneID(record.ID).
		SetValue(setting.Value).
		ClearDeletedAt().
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainAccountSetting(updated), nil
}

func toDomainAccountSetting(record *entclient.AccountSetting) *domainaccountsetting.AccountSetting {
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

func toDomainAccountSettings(records []*entclient.AccountSetting) []*domainaccountsetting.AccountSetting {
	settings := make([]*domainaccountsetting.AccountSetting, 0, len(records))
	for _, record := range records {
		settings = append(settings, toDomainAccountSetting(record))
	}
	return settings
}
