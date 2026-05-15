package account

import "context"

type Repository interface {
	GetByWechatOpenID(ctx context.Context, wechatOpenID string) (*Account, error)
	Create(ctx context.Context, account *Account) (*Account, error)
	Update(ctx context.Context, account *Account) (*Account, error)
	Delete(ctx context.Context, id int) error
}
