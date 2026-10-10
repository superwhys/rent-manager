package app

import (
	"context"
	"strings"

	"github.com/superwhys/rent-manager/internal/xerr"
)

type accountIDContextKey struct{}

// EmptyRequest 表示不需要绑定参数的请求。
type EmptyRequest struct{}

// Pagination 表示通用分页查询参数。
type Pagination struct {
	Limit  int `form:"limit"  json:"limit"`
	Offset int `form:"offset" json:"offset"`
}

// AccountIDFromContext 从上下文中的令牌声明获取当前账号 ID。
func AccountIDFromContext(ctx context.Context) (string, error) {
	accountID, ok := ctx.Value(accountIDContextKey{}).(string)
	if !ok || strings.TrimSpace(accountID) == "" {
		return "", xerr.ErrUnauthorized
	}
	return accountID, nil
}

// SetAccountIDToContext 将当前账号 ID 写入上下文。
func SetAccountIDToContext(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDContextKey{}, accountID)
}
