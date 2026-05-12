package account

import (
	"strings"
	"time"
)

// Organization groups buildings and employee accounts.
type Organization struct {
	ID        int
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewOrganization(params CreateOrganizationParams) (*Organization, error) {
	organization := &Organization{
		Name:     strings.TrimSpace(params.Name),
		IsActive: true,
	}
	if err := organization.Validate(); err != nil {
		return nil, err
	}
	return organization, nil
}

func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrOrganizationNameRequired
	}
	return nil
}
