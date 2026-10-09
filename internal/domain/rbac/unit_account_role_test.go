package rbac

import (
	"errors"
	"testing"
)

func TestUnitAccountRoleUsesAuthkitAccountIDs(t *testing.T) {
	ownerID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	targetID := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	grant, err := NewUnitAccountRole(UnitAccountRolesParams{
		OwnerAccountID:  ownerID,
		TargetAccountID: targetID,
		UnitID:          7,
	}, 3)
	if err != nil {
		t.Fatalf("NewUnitAccountRole() error = %v", err)
	}
	if grant.AccountID != targetID || grant.GrantedByAccountID != ownerID {
		t.Fatalf("grant = %#v, want unchanged authkit account IDs", grant)
	}
}

func TestUnitAccountRoleRejectsMissingAccountsAndSelfGrant(t *testing.T) {
	ownerID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	targetID := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	tests := []struct {
		name   string
		owner  string
		target string
		want   error
	}{
		{name: "missing owner", target: targetID, want: ErrRoleAccountRequired},
		{name: "blank owner", owner: " \t", target: targetID, want: ErrRoleAccountRequired},
		{name: "missing target", owner: ownerID, want: ErrTargetAccountRequired},
		{name: "blank target", owner: ownerID, target: " \t", want: ErrTargetAccountRequired},
		{name: "self grant", owner: ownerID, target: ownerID, want: ErrCannotGrantToOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUnitAccountRole(UnitAccountRolesParams{
				OwnerAccountID:  tt.owner,
				TargetAccountID: tt.target,
				UnitID:          7,
			}, 3)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewUnitAccountRole() error = %v, want %v", err, tt.want)
			}
		})
	}
}
