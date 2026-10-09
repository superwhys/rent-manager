package rbac

import (
	"strings"
	"time"
)

const (
	UnitGrantStatusActive   = "active"
	UnitGrantStatusDisabled = "disabled"
)

type UnitAccountRole struct {
	ID                 int
	UnitID             int
	AccountID          string
	RoleID             int
	GrantedByAccountID string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
	Role               *Role
}

func NewUnitAccountRole(params UnitAccountRolesParams, roleID int) (*UnitAccountRole, error) {
	grant := &UnitAccountRole{
		UnitID:             params.UnitID,
		AccountID:          params.TargetAccountID,
		RoleID:             roleID,
		GrantedByAccountID: params.OwnerAccountID,
		Status:             UnitGrantStatusActive,
	}
	if err := grant.Validate(); err != nil {
		return nil, err
	}
	return grant, nil
}

func (g *UnitAccountRole) Validate() error {
	if strings.TrimSpace(g.GrantedByAccountID) == "" {
		return ErrRoleAccountRequired
	}
	if g.UnitID <= 0 {
		return ErrUnitRequired
	}
	if strings.TrimSpace(g.AccountID) == "" {
		return ErrTargetAccountRequired
	}
	if g.AccountID == g.GrantedByAccountID {
		return ErrCannotGrantToOwner
	}
	if g.RoleID <= 0 {
		return ErrRoleNotFound
	}
	switch g.Status {
	case UnitGrantStatusActive, UnitGrantStatusDisabled:
		return nil
	default:
		return ErrUnitGrantNotFound
	}
}
