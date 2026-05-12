package building

import (
	"strings"
	"time"
)

type Building struct {
	ID                  int
	OrganizationID      int
	Name                string
	Address             string
	City                string
	Province            string
	DefaultMeterItemIDs []int
	DefaultMeterItems   []*MeterItem
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type MeterItem struct {
	ID       int
	Code     string
	Unit     string
	IsSystem bool
	CalcType string
	Remark   *string
}

func NewBuilding(params CreateParams) (*Building, error) {
	building := &Building{
		OrganizationID:      params.OrganizationID,
		Name:                strings.TrimSpace(params.Name),
		Address:             strings.TrimSpace(params.Address),
		City:                strings.TrimSpace(params.City),
		Province:            strings.TrimSpace(params.Province),
		DefaultMeterItemIDs: normalizeMeterItemIDs(params.DefaultMeterItemIDs),
	}
	if err := building.Validate(); err != nil {
		return nil, err
	}
	return building, nil
}

func (b *Building) Update(params UpdateParams) error {
	b.ID = params.ID
	b.OrganizationID = params.OrganizationID
	b.Name = strings.TrimSpace(params.Name)
	b.Address = strings.TrimSpace(params.Address)
	b.City = strings.TrimSpace(params.City)
	b.Province = strings.TrimSpace(params.Province)
	b.DefaultMeterItemIDs = normalizeMeterItemIDs(params.DefaultMeterItemIDs)
	return b.Validate()
}

func (b *Building) Validate() error {
	if b.OrganizationID <= 0 {
		return ErrOrganizationRequired
	}
	if b.Name == "" {
		return ErrNameRequired
	}
	if b.Address == "" {
		return ErrAddressRequired
	}
	if b.City == "" {
		return ErrCityRequired
	}
	if b.Province == "" {
		return ErrProvinceRequired
	}
	for _, id := range b.DefaultMeterItemIDs {
		if id <= 0 {
			return ErrInvalidMeterItemID
		}
	}
	return nil
}

func normalizeMeterItemIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}
