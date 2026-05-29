package services

import (
	"context"
	"strconv"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

const (
	unitOccupancyStatusVacancy = "vacancy"
	unitOccupancyStatusFull    = "full"
)

// UnitService 提供单元管理用例。
type UnitService struct {
	repo     domainunit.Repository
	roomRepo domainroom.Repository
	rbac     *RBACService
}

// NewUnitService 创建单元管理服务。
func NewUnitService(repo domainunit.Repository, roomRepo domainroom.Repository, rbac *RBACService) *UnitService {
	return &UnitService{repo: repo, roomRepo: roomRepo, rbac: rbac}
}

// CreateUnit 创建当前账号下的单元。
func (s *UnitService) CreateUnit(ctx context.Context, cmd *requests.CreateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := domainunit.NewUnit(domainunit.CreateParams{
		AccountID:           accountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit)
}

// GetUnit 获取当前账号可访问的单元详情。
func (s *UnitService) GetUnit(ctx context.Context, id int) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, id, domainrbac.PermissionUnitView)
	if err != nil {
		return nil, err
	}

	return unit, nil
}

// ListUnits 获取当前账号可访问的单元列表。
func (s *UnitService) ListUnits(ctx context.Context, query *requests.ListUnitsQuery) ([]*domainunit.Summary, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scope := query.Scope
	if scope == "" {
		scope = domainunit.ListScopeAll
	}

	switch scope {
	case domainunit.ListScopeOwned, domainunit.ListScopeGranted, domainunit.ListScopeAll:
	default:
		return nil, domainunit.ErrInvalidListScope
	}

	units, err := s.repo.List(ctx, domainunit.ListQuery{
		AccountID: accountID,
		Scope:     scope,
	})
	if err != nil {
		return nil, err
	}

	return s.buildUnitSummaries(ctx, units)
}

// UpdateUnit 更新当前账号可管理的单元。
func (s *UnitService) UpdateUnit(ctx context.Context, cmd *requests.UpdateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.requireUnitPermission(ctx, accountID, cmd.UnitID, domainrbac.PermissionUnitUpdate)
	if err != nil {
		return nil, err
	}

	unit := &domainunit.Unit{}
	if err := unit.Update(domainunit.UpdateParams{
		ID:                  cmd.UnitID,
		AccountID:           existing.AccountID, // 保持创建人不变
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, unit)
	if err != nil {
		return nil, err
	}

	setUnitAccessType(updated, accountID)
	return updated, nil
}

// DeleteUnit 删除当前账号拥有的单元。
func (s *UnitService) DeleteUnit(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}

func (s *UnitService) requireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (*domainunit.Unit, error) {
	unit, err := s.repo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
		return unit, nil
	}

	if err := s.rbac.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		return nil, err
	}

	unit.AccessType = domainunit.AccessTypeGranted
	return unit, nil
}

func (s *UnitService) buildUnitSummaries(ctx context.Context, units []*domainunit.Unit) ([]*domainunit.Summary, error) {
	summaries := make([]*domainunit.Summary, 0, len(units))
	active := true

	for _, unit := range units {
		if unit == nil {
			continue
		}

		roomCount, err := s.roomRepo.Count(ctx, unit.AccountID, domainroom.ListQuery{
			UnitID: unit.ID,
		})
		if err != nil {
			return nil, err
		}
		rentedCount, err := s.roomRepo.Count(ctx, unit.AccountID, domainroom.ListQuery{
			UnitID:   unit.ID,
			IsActive: &active,
		})
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, &domainunit.Summary{
			Unit:              unit,
			RoomCount:         roomCount,
			RentedCount:       rentedCount,
			MonthlyRentAmount: mockUnitMonthlyRentAmount(unit),
			OccupancyStatus:   unitOccupancyStatus(roomCount, rentedCount),
		})
	}

	return summaries, nil
}

func setUnitAccessType(unit *domainunit.Unit, accountID int) {
	if unit == nil {
		return
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}
}

func unitOccupancyStatus(roomCount int, rentedCount int) string {
	if roomCount > 0 && rentedCount >= roomCount {
		return unitOccupancyStatusFull
	}

	return unitOccupancyStatusVacancy
}

func mockUnitMonthlyRentAmount(unit *domainunit.Unit) string {
	if unit == nil {
		return "0"
	}

	amount := 6000 + unit.ID*800

	return formatAmountWithComma(amount)
}

func formatAmountWithComma(amount int) string {
	text := strconv.Itoa(amount)
	if len(text) <= 3 {
		return text
	}

	prefixLen := len(text) % 3
	if prefixLen == 0 {
		prefixLen = 3
	}

	formatted := text[:prefixLen]
	for i := prefixLen; i < len(text); i += 3 {
		formatted += "," + text[i:i+3]
	}

	return formatted
}
