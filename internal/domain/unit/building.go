package unit

import (
	"strings"
	"time"
)

type Unit struct {
	ID                  int
	AccountID           int
	Name                string
	Address             string
	City                string
	Province            string
	AccessType          string
	DefaultMeterItemIDs []int
	DefaultMeterItems   []*MeterItem
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Summary struct {
	Unit              *Unit
	RoomCount         int
	RentedCount       int
	MonthlyRentAmount string
	OccupancyStatus   string
}

type MeterItem struct {
	ID        int
	Code      string
	Unit      string
	UnitPrice int
	IsSystem  bool
	CalcType  string
	Remark    *string
}

func NewUnit(params CreateParams) (*Unit, error) {
	unit := &Unit{
		AccountID:           params.AccountID,
		Name:                strings.TrimSpace(params.Name),
		Address:             strings.TrimSpace(params.Address),
		City:                strings.TrimSpace(params.City),
		Province:            strings.TrimSpace(params.Province),
		DefaultMeterItemIDs: normalizeMeterItemIDs(params.DefaultMeterItemIDs),
	}
	if err := unit.Validate(); err != nil {
		return nil, err
	}
	return unit, nil
}

func (u *Unit) Update(params UpdateParams) error {
	u.ID = params.ID
	u.AccountID = params.AccountID
	u.Name = strings.TrimSpace(params.Name)
	u.Address = strings.TrimSpace(params.Address)
	u.City = strings.TrimSpace(params.City)
	u.Province = strings.TrimSpace(params.Province)
	u.DefaultMeterItemIDs = normalizeMeterItemIDs(params.DefaultMeterItemIDs)
	return u.Validate()
}

func (u *Unit) Validate() error {
	if u.AccountID <= 0 {
		return ErrAccountRequired
	}
	if u.Name == "" {
		return ErrNameRequired
	}
	if u.Address == "" {
		return ErrAddressRequired
	}
	if u.City == "" {
		return ErrCityRequired
	}
	if u.Province == "" {
		return ErrProvinceRequired
	}
	for _, id := range u.DefaultMeterItemIDs {
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
