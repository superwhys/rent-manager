package home

import (
	"context"
	"time"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainrentbill "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rentbill"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
)

// Service 提供首页聚合数据用例。
type Service struct {
	unitRepo     domainunit.Repository
	roomRepo     domainroom.Repository
	rentBillRepo domainrentbill.Repository
}

// NewService 创建首页聚合服务。
func NewService(unitRepo domainunit.Repository, roomRepo domainroom.Repository, rentBillRepos ...domainrentbill.Repository) *Service {
	var rentBillRepo domainrentbill.Repository
	if len(rentBillRepos) > 0 {
		rentBillRepo = rentBillRepos[0]
	}
	return &Service{
		unitRepo:     unitRepo,
		roomRepo:     roomRepo,
		rentBillRepo: rentBillRepo,
	}
}

// GetHomeSummary 获取当前账号可访问的首页楼栋与房间汇总。
func (s *Service) GetHomeSummary(ctx context.Context) (*Summary, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	units, err := s.unitRepo.List(ctx, domainunit.ListQuery{
		AccountID: accountID,
		Scope:     domainunit.ListScopeAll,
	})
	if err != nil {
		return nil, err
	}

	currentMonth, previousMonth := s.currentAndPreviousMonth()
	summary := &Summary{CurrentMonth: currentMonth}
	active := true
	inactive := false
	for _, unit := range units {
		if unit == nil {
			continue
		}
		summary.BuildingCount++

		// TODO: 当前尚未建租约模型，出租/空置沿用房间启用状态统计。
		rentedCount, err := s.countRoomsInUnit(ctx, unit, &active)
		if err != nil {
			return nil, err
		}
		vacantCount, err := s.countRoomsInUnit(ctx, unit, &inactive)
		if err != nil {
			return nil, err
		}

		summary.RentedRoomCount += rentedCount
		summary.VacantRoomCount += vacantCount
		summary.RoomCount += rentedCount + vacantCount
		if err := s.collectRentBillSummary(ctx, unit, summary, currentMonth, previousMonth); err != nil {
			return nil, err
		}
	}
	// TODO:
	summary.MonthlyIncomeGrowthCents = summary.CurrentMonthIncomeCents - summary.PreviousMonthIncomeCents

	return summary, nil
}

// countRoomsInUnit 统计指定楼栋下符合状态条件的房间数量。
func (s *Service) countRoomsInUnit(ctx context.Context, unit *domainunit.Unit, isActive *bool) (int, error) {
	return s.roomRepo.Count(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: isActive,
	})
}

// collectRentBillSummary 汇总指定楼栋当前月和上月的租金账单数据。
func (s *Service) collectRentBillSummary(ctx context.Context, unit *domainunit.Unit, summary *Summary, currentMonth string, previousMonth string) error {
	if s.rentBillRepo == nil {
		return nil
	}
	currentIncome, _, err := s.sumRentBills(ctx, unit, currentMonth, domainrentbill.StatusPaid)
	if err != nil {
		return err
	}
	previousIncome, _, err := s.sumRentBills(ctx, unit, previousMonth, domainrentbill.StatusPaid)
	if err != nil {
		return err
	}
	_, unpaidCount, err := s.sumRentBills(ctx, unit, currentMonth, domainrentbill.StatusUnpaid)
	if err != nil {
		return err
	}
	summary.CurrentMonthIncomeCents += currentIncome
	summary.PreviousMonthIncomeCents += previousIncome
	summary.CurrentMonthUnpaidBillCount += unpaidCount
	return nil
}

// sumRentBills 汇总指定楼栋、月份和状态下的账单金额与数量。
func (s *Service) sumRentBills(ctx context.Context, unit *domainunit.Unit, periodMonth string, status string) (int, int, error) {
	bills, err := s.rentBillRepo.List(ctx, unit.AccountID, domainrentbill.ListQuery{
		UnitID:      unit.ID,
		PeriodMonth: periodMonth,
		Status:      status,
	})
	if err != nil {
		return 0, 0, err
	}
	totalAmount := 0
	count := 0
	for _, bill := range bills {
		if bill == nil {
			continue
		}
		totalAmount += bill.TotalAmount
		count++
	}
	return totalAmount, count, nil
}

// currentAndPreviousMonth 返回当前月份和上一月份的 YYYY-MM 表示。
func (s *Service) currentAndPreviousMonth() (string, string) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return monthStart.Format("2006-01"), monthStart.AddDate(0, -1, 0).Format("2006-01")
}
