package rbac

import (
	"strings"
	"time"
)

type Role struct {
	ID             int
	AccountID      int
	Name           string
	IsDefaultOwner bool
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func NewRole(params CreateRoleParams) (*Role, error) {
	role := &Role{
		AccountID:      params.AccountID,
		Name:           strings.TrimSpace(params.Name),
		IsDefaultOwner: params.IsDefaultOwner,
		IsActive:       true,
	}
	if err := role.Validate(); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *Role) Validate() error {
	if r.AccountID <= 0 {
		return ErrRoleAccountRequired
	}
	if r.Name == "" {
		return ErrRoleNameRequired
	}
	return nil
}
