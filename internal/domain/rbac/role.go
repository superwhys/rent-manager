package rbac

import (
	"strings"
	"time"
)

type Role struct {
	ID        int
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewRole(params CreateRoleParams) (*Role, error) {
	role := &Role{
		Name:     strings.TrimSpace(params.Name),
		IsActive: true,
	}
	if err := role.Validate(); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *Role) Validate() error {
	if r.Name == "" {
		return ErrRoleNameRequired
	}
	return nil
}
