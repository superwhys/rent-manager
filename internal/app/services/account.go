package services

import (
	"context"

	domainaccount "github.com/superwhys/rent-manager-server/internal/domain/account"
)

type AccountService struct {
	repo domainaccount.Repository
}

func NewAccountService(repo domainaccount.Repository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Self(ctx context.Context) (*domainaccount.Account, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, accountID)
}
