package account

import (
	"context"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainaccount "github.com/superwhys/rent-manager/internal/domain/account"
)

// Service 提供账号相关用例。
type Service struct {
	repo domainaccount.Repository
}

// NewService 创建账号服务。
func NewService(repo domainaccount.Repository) *Service {
	return &Service{repo: repo}
}

// Self 获取当前登录账号信息。
func (s *Service) Self(ctx context.Context) (*domainaccount.Account, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, accountID)
}
