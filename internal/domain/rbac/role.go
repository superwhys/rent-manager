package rbac

import (
	"strings"
	"time"
)

type Role struct {
	ID             int
	AccountID      string
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
	if strings.TrimSpace(r.AccountID) == "" {
		return ErrRoleAccountRequired
	}
	if r.Name == "" {
		return ErrRoleNameRequired
	}
	return nil
}

func (r *Role) Update(params UpdateRoleParams) error {
	r.ID = params.ID
	r.AccountID = params.AccountID
	r.Name = strings.TrimSpace(params.Name)
	r.IsDefaultOwner = params.IsDefaultOwner
	r.IsActive = params.IsActive
	return r.Validate()
}
