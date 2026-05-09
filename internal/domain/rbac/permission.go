package rbac

import (
	"strings"
	"time"
)

type Permission struct {
	ID          int
	Code        string
	APIPath     string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewPermission(params CreatePermissionParams) (*Permission, error) {
	permission := &Permission{
		Code:        strings.TrimSpace(params.Code),
		APIPath:     strings.TrimSpace(params.APIPath),
		Description: strings.TrimSpace(params.Description),
		IsActive:    true,
	}
	if err := permission.Validate(); err != nil {
		return nil, err
	}
	return permission, nil
}

func (p *Permission) Validate() error {
	if p.Code == "" {
		return ErrPermissionCodeRequired
	}
	if p.APIPath == "" {
		return ErrPermissionAPIPathRequired
	}
	if p.Description == "" {
		return ErrPermissionDescriptionRequired
	}
	return nil
}
