package services

import (
	"context"
	"testing"

	domainaccount "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/account"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/jwtutils"
)

func TestAccountServiceSelf(t *testing.T) {
	repo := &fakeAccountRepository{
		accounts: map[int]*domainaccount.Account{
			1: {ID: 1, Name: "test", Email: "test@example.com", WechatOpenID: "openid", IsActive: true},
		},
	}
	service := NewAccountService(repo)

	got, err := service.Self(jwtutils.SetTokenClaimsToContext(context.Background(), &jwtutils.AccountTokenClaims{UserID: 1}))
	if err != nil {
		t.Fatalf("Self() error = %v", err)
	}
	if got.ID != 1 || got.Name != "test" {
		t.Fatalf("unexpected account: %#v", got)
	}
}

type fakeAccountRepository struct {
	accounts map[int]*domainaccount.Account
}

func (f *fakeAccountRepository) Get(_ context.Context, id int) (*domainaccount.Account, error) {
	account, ok := f.accounts[id]
	if !ok {
		return nil, domainaccount.ErrAccountNotFound
	}
	return account, nil
}

func (f *fakeAccountRepository) GetByWechatOpenID(context.Context, string) (*domainaccount.Account, error) {
	return nil, domainaccount.ErrAccountNotFound
}

func (f *fakeAccountRepository) Create(_ context.Context, account *domainaccount.Account) (*domainaccount.Account, error) {
	return account, nil
}

func (f *fakeAccountRepository) Update(_ context.Context, account *domainaccount.Account) (*domainaccount.Account, error) {
	return account, nil
}

func (f *fakeAccountRepository) Delete(context.Context, int) error {
	return nil
}
