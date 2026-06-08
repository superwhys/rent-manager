package app

import (
	"context"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

type accountIDContextKey struct{}

// EmptyRequest 表示不需要绑定参数的请求。
type EmptyRequest struct{}

// Pagination 表示通用分页查询参数。
type Pagination struct {
	Limit  int `form:"limit" json:"limit"`
	Offset int `form:"offset" json:"offset"`
}

// AccountIDFromContext 从上下文中的令牌声明获取当前账号 ID。
func AccountIDFromContext(ctx context.Context) (int, error) {
	accountID, ok := ctx.Value(accountIDContextKey{}).(int)
	if !ok || accountID <= 0 {
		return 0, xerr.ErrUnauthorized
	}
	return accountID, nil
}

// SetAccountIDToContext 将当前账号 ID 写入上下文。
func SetAccountIDToContext(ctx context.Context, accountID int) context.Context {
	return context.WithValue(ctx, accountIDContextKey{}, accountID)
}
