package app

import (
	"context"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/jwtutils"
)

// EmptyRequest 表示不需要绑定参数的请求。
type EmptyRequest struct{}

// Pagination 表示通用分页查询参数。
type Pagination struct {
	Limit  int `form:"limit" json:"limit"`
	Offset int `form:"offset" json:"offset"`
}

// AccountIDFromContext 从上下文中的令牌声明获取当前账号 ID。
func AccountIDFromContext(ctx context.Context) (int, error) {
	claims, err := jwtutils.TokenClaimsFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
