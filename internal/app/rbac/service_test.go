package rbac

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/miebyte/authkit"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestGrantUnitRolesUsesAuthkitAccountAfterOwnershipCheck(t *testing.T) {
	ownerID, targetID := strings.Repeat("a", 64), strings.Repeat("b", 64)
	storageErr := errors.New("身份仓储不可用")
	for _, test := range []struct {
		name          string
		unitErr       error
		accountErr    error
		wantErr       error
		lookupAccount bool
		grantRoles    bool
	}{
		{name: "authkit 账号直接授权", lookupAccount: true, grantRoles: true},
		{name: "外部账号不存在", accountErr: authkit.ErrNotFound, wantErr: domainrbac.ErrAccountNotFound, lookupAccount: true},
		{name: "身份仓储错误直接返回", accountErr: storageErr, wantErr: storageErr, lookupAccount: true},
		{name: "跨账号单元不查询身份", unitErr: domainunit.ErrNotFound, wantErr: domainrbac.ErrUnitNotInAccount},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &grantRepository{}
			accounts := &accountLookup{err: test.accountErr}
			units := &ownedUnitLookup{err: test.unitErr}
			service := NewService(repo, accounts, units, nil)
			ctx := appcore.SetAccountIDToContext(context.Background(), ownerID)
			_, err := service.GrantUnitRoles(ctx, &UnitAccountRolesCommand{UnitID: 7, AccountID: targetID, RoleIDs: []int{3, 3}})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GrantUnitRoles error = %v, want %v", err, test.wantErr)
			}
			if units.accountID != ownerID || units.unitID != 7 {
				t.Fatalf("ownership lookup = %+v", units)
			}
			if (accounts.id != "") != test.lookupAccount || accounts.id != "" && accounts.id != targetID {
				t.Fatalf("authkit lookup ID = %q", accounts.id)
			}
			if repo.called != test.grantRoles {
				t.Fatalf("grant repository called = %v, want %v", repo.called, test.grantRoles)
			}
			if test.grantRoles && (repo.params.OwnerAccountID != ownerID || repo.params.TargetAccountID != targetID || len(repo.params.RoleIDs) != 1) {
				t.Fatalf("grant must preserve authkit IDs: %+v", repo.params)
			}
		})
	}
}

type accountLookup struct {
	id  string
	err error
}

func (a *accountLookup) GetByID(_ context.Context, id string) (*authkit.Account, error) {
	a.id = id
	if a.err != nil {
		return nil, a.err
	}
	return &authkit.Account{ID: id}, nil
}

type ownedUnitLookup struct {
	accountID string
	unitID    int
	err       error
}

func (u *ownedUnitLookup) Get(_ context.Context, accountID string, unitID int) (*domainunit.Unit, error) {
	u.accountID, u.unitID = accountID, unitID
	return &domainunit.Unit{ID: unitID, AccountID: accountID}, u.err
}

type grantRepository struct {
	domainrbac.Repository
	called bool
	params domainrbac.UnitAccountRolesParams
}

func (r *grantRepository) GrantUnitRoles(_ context.Context, params domainrbac.UnitAccountRolesParams) ([]*domainrbac.UnitAccountRole, error) {
	r.called, r.params = true, params
	return nil, nil
}
