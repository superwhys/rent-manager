package accountsetting

import "context"

type Repository interface {
	GetByKey(ctx context.Context, accountID int, key string) (*AccountSetting, error)
	List(ctx context.Context, query ListQuery) ([]*AccountSetting, error)
	Upsert(ctx context.Context, setting *AccountSetting) (*AccountSetting, error)
}
