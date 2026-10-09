package unit

import (
	"context"
	"strconv"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

const (
	unitOccupancyStatusVacancy = "vacancy"
	unitOccupancyStatusFull    = "full"
)

// Service 提供单元管理用例。
type Service struct {
	repo        domainunit.Repository
	roomRepo    domainroom.Repository
	rbacService domainrbac.IRbacService
}

// NewService 创建单元管理服务。
func NewService(repo domainunit.Repository, roomRepo domainroom.Repository, rbacRepo domainrbac.Repository) *Service {
	return &Service{
		repo:        repo,
		roomRepo:    roomRepo,
		rbacService: domainrbac.NewRbacService(rbacRepo),
	}
}

// CreateUnit 创建当前账号下的单元。
func (s *Service) CreateUnit(ctx context.Context, cmd *CreateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	config, err := toDomainUnitConfig(cmd.JsonExtend)
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
		Config:              config,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit)
}

// GetUnit 获取当前账号可访问的单元详情。
func (s *Service) GetUnitInfo(ctx context.Context, id int) (*domainunit.Unit, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, id, domainrbac.PermissionUnitView)
	if err != nil {
		return nil, err
	}

	// TODO:
	// 1. 获取预估月收入
	// 2. 获取房间数
	// 3. 获取在租数（可以租的）

	return unit, nil
}

// ListUnits 获取当前账号可访问的单元列表。
func (s *Service) ListUnits(ctx context.Context, query *ListUnitsQuery) ([]*domainunit.Summary, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
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
func (s *Service) UpdateUnit(ctx context.Context, cmd *UpdateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.requireUnitPermission(ctx, accountID, cmd.UnitID, domainrbac.PermissionUnitUpdate)
	if err != nil {
		return nil, err
	}
	config, err := toDomainUnitConfig(cmd.JsonExtend)
	if err != nil {
		return nil, err
	}

	unit := &domainunit.Unit{}
	if err := unit.Update(domainunit.UpdateParams{
		ID:                  cmd.UnitID,
		AccountID:           existing.AccountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
		Config:              config,
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
func (s *Service) DeleteUnit(ctx context.Context, id int) error {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}

// requireUnitPermission 校验当前账号对单元的权限并标记访问类型。
func (s *Service) requireUnitPermission(ctx context.Context, accountID string, unitID int, permissionCode string) (*domainunit.Unit, error) {
	err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, permissionCode)
	if err != nil {
		return nil, err
	}

	unit, err := s.repo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}

	return unit, nil
}

// buildUnitSummaries 汇总单元列表的房间统计信息。
func (s *Service) buildUnitSummaries(ctx context.Context, units []*domainunit.Unit) ([]*domainunit.Summary, error) {
	summaries := make([]*domainunit.Summary, 0, len(units))
	rented := false

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
			IsActive: &rented,
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

// setUnitAccessType 根据账号与单元归属设置访问类型。
func setUnitAccessType(unit *domainunit.Unit, accountID string) {
	if unit == nil {
		return
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}
}

// unitOccupancyStatus 根据房间数量和已租数量计算出租状态。
func unitOccupancyStatus(roomCount int, rentedCount int) string {
	if roomCount > 0 && rentedCount >= roomCount {
		return unitOccupancyStatusFull
	}

	return unitOccupancyStatusVacancy
}

// mockUnitMonthlyRentAmount 返回当前阶段用于首页展示的月租金金额。
func mockUnitMonthlyRentAmount(unit *domainunit.Unit) string {
	if unit == nil {
		return "0"
	}

	amount := 6000 + unit.ID*800

	return formatAmountWithComma(amount)
}

// formatAmountWithComma 将整数金额格式化为千分位字符串。
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
