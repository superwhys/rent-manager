package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

// AccountProfileRepository 使用 authkit 的账号模型更新展示资料。
type AccountProfileRepository struct {
	q *query.Query
}

func NewAccountProfileRepository(client *gorm.DB) *AccountProfileRepository {
	return &AccountProfileRepository{q: query.Use(client)}
}

func (r *AccountProfileRepository) UpdateUsername(ctx context.Context, accountID, username string) error {
	a := r.q.Account
	_, err := a.WithContext(ctx).Where(a.ID.Eq(accountID)).Update(a.Username, username)
	return err
}
