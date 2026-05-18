package rbac

import "time"

const (
	UnitGrantStatusActive   = "active"
	UnitGrantStatusDisabled = "disabled"
)

type UnitAccountRole struct {
	ID                 int
	UnitID             int
	AccountID          int
	RoleID             int
	GrantedByAccountID int
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
	if g.GrantedByAccountID <= 0 {
		return ErrRoleAccountRequired
	}
	if g.UnitID <= 0 {
		return ErrUnitRequired
	}
	if g.AccountID <= 0 {
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
