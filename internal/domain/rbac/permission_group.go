package rbac

import (
	"strings"
	"time"
)

// PermissionGroup groups global permissions for organization roles.
type PermissionGroup struct {
	ID          int
	Code        string
	Name        string
	Description *string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewPermissionGroup(params CreatePermissionGroupParams) (*PermissionGroup, error) {
	description := strings.TrimSpace(params.Description)
	group := &PermissionGroup{
		Code:      strings.TrimSpace(params.Code),
		Name:      strings.TrimSpace(params.Name),
		SortOrder: params.SortOrder,
		IsActive:  true,
	}
	if description != "" {
		group.Description = &description
	}
	if err := group.Validate(); err != nil {
		return nil, err
	}
	return group, nil
}

func (g *PermissionGroup) Validate() error {
	if g.Code == "" {
		return ErrPermissionGroupCodeRequired
	}
	if g.Name == "" {
		return ErrPermissionGroupNameRequired
	}
	return nil
}
