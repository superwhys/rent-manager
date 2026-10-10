package home

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

const (
	maxHomeTodos = 10

	todoTagMeterReading   = "抄表"
	todoTagBillGeneration = "账单"
	todoTagRentCollection = "收租"
)

var homeLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

// Service 提供首页聚合数据用例。
type Service struct {
	unitRepo         domainunit.Repository
	roomRepo         domainroom.Repository
	rentBillRepo     domainrentbill.Repository
	meterReadingRepo domainmeterreading.Repository
	rbacService      domainrbac.IRbacService
	now              func() time.Time
}

// NewService 创建首页聚合服务。
func NewService(
	unitRepo domainunit.Repository,
	roomRepo domainroom.Repository,
	rentBillRepo domainrentbill.Repository,
	meterReadingRepo domainmeterreading.Repository,
	rbacRepo domainrbac.Repository,
) *Service {
	var rbacService domainrbac.IRbacService
	if rbacRepo != nil {
		rbacService = domainrbac.NewRbacService(rbacRepo)
	}
	return &Service{
		unitRepo:         unitRepo,
		roomRepo:         roomRepo,
		rentBillRepo:     rentBillRepo,
		meterReadingRepo: meterReadingRepo,
		rbacService:      rbacService,
		now:              time.Now,
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

	now := s.currentTime().In(homeLocation)
	currentMonth, previousMonth := currentAndPreviousMonth(now)
	summary := &Summary{CurrentMonth: currentMonth}
	todoBuckets := &homeTodoBuckets{}

	for _, unit := range units {
		if unit == nil {
			continue
		}
		summary.BuildingCount++
		canViewRooms, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionRoomView)
		if err != nil {
			return nil, err
		}
		canViewRentBill, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionRentBillView)
		if err != nil {
			return nil, err
		}

		// 查询出所有已出租的房间
		rentedRooms, err := s.listRoomsInUnit(ctx, unit, new(false))
		if err != nil {
			return nil, err
		}

		if canViewRooms {
			availableCount, err := s.countRoomsInUnit(ctx, unit, new(true))
			if err != nil {
				return nil, err
			}
			summary.RentedRoomCount += len(rentedRooms)
			summary.VacantRoomCount += availableCount
			summary.RoomCount += len(rentedRooms) + availableCount
		}
		if canViewRentBill {
			if err := s.collectRentBillSummary(ctx, unit, summary, currentMonth, previousMonth); err != nil {
				return nil, err
			}
		}
		unitTodos, err := s.collectUnitTodos(ctx, accountID, unit, rentedRooms, now)
		if err != nil {
			return nil, err
		}
		todoBuckets.append(unitTodos)
	}

	summary.MonthlyIncomeGrowthCents = summary.CurrentMonthIncomeCents - summary.PreviousMonthIncomeCents
	summary.Todos = todoBuckets.todos(maxHomeTodos)

	return summary, nil
}

// listRoomsInUnit 查询指定楼栋下符合状态条件的房间。
func (s *Service) listRoomsInUnit(
	ctx context.Context,
	unit *domainunit.Unit,
	isActive *bool,
) ([]*domainroom.Room, error) {
	return s.roomRepo.List(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: isActive,
	})
}

// countRoomsInUnit 统计指定楼栋下符合状态条件的房间数量。
func (s *Service) countRoomsInUnit(
	ctx context.Context,
	unit *domainunit.Unit,
	isActive *bool,
) (int, error) {
	return s.roomRepo.Count(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: isActive,
	})
}

// collectRentBillSummary 汇总指定楼栋当前月和上月的租金账单数据。
func (s *Service) collectRentBillSummary(
	ctx context.Context,
	unit *domainunit.Unit,
	summary *Summary,
	currentMonth string,
	previousMonth string,
) error {
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
func (s *Service) sumRentBills(
	ctx context.Context,
	unit *domainunit.Unit,
	periodMonth string,
	status string,
) (int, int, error) {
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

func (s *Service) collectUnitTodos(
	ctx context.Context,
	accountID string,
	unit *domainunit.Unit,
	rentedRooms []*domainroom.Room,
	now time.Time,
) (*unitTodoResult, error) {
	result := &unitTodoResult{}
	currentMonth := now.Format("2006-01")
	canViewMeterReading, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionMeterReadingView)
	if err != nil {
		return nil, err
	}
	canViewRentBill, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionRentBillView)
	if err != nil {
		return nil, err
	}
	if !canViewMeterReading && !canViewRentBill {
		return result, nil
	}

	readingsByPeriod, err := s.listUnitReadingsByPeriod(ctx, unit, currentMonth)
	if err != nil {
		return nil, err
	}
	periodSet := map[string]struct{}{currentMonth: {}}
	for month := range readingsByPeriod {
		periodSet[month] = struct{}{}
	}

	billRoomIDsByPeriod := make(map[string]map[int]bool)
	unpaidCountByPeriod := make(map[string]int)
	if s.rentBillRepo != nil {
		bills, err := s.rentBillRepo.List(ctx, unit.AccountID, domainrentbill.ListQuery{UnitID: unit.ID})
		if err != nil {
			return nil, err
		}
		for _, bill := range bills {
			if bill == nil {
				continue
			}
			if billRoomIDsByPeriod[bill.PeriodMonth] == nil {
				billRoomIDsByPeriod[bill.PeriodMonth] = make(map[int]bool)
			}
			billRoomIDsByPeriod[bill.PeriodMonth][bill.RoomID] = true
			if canViewRentBill && bill.Status == domainrentbill.StatusUnpaid {
				unpaidCountByPeriod[bill.PeriodMonth]++
			}
		}
	}

	periods := make([]string, 0, len(periodSet))
	for month := range periodSet {
		periods = append(periods, month)
	}
	sort.Strings(periods)
	for _, month := range periods {
		missingReadingCount, generationCount := 0, 0
		for _, room := range rentedRooms {
			if room == nil {
				continue
			}
			roomReadings := readingsByPeriod[month][room.ID]
			hasBill := billRoomIDsByPeriod[month][room.ID]
			complete, hasMeterItems := isRoomMeterReadingComplete(room, roomReadings)
			// 历史月份只追踪已有抄表记录，不推测过去的出租状态。
			needsReading := month == currentMonth || len(roomReadings) > 0
			if canViewMeterReading && needsReading && hasMeterItems && !complete && (month == currentMonth || !hasBill) {
				missingReadingCount++
			}
			if canViewRentBill && s.rentBillRepo != nil && needsReading && complete && !hasBill {
				generationCount++
			}
		}
		if missingReadingCount > 0 {
			todo := *newHomeTodo(unit.Name, meterReadingTodoDescription(unit, month, missingReadingCount, now), todoTagMeterReading)
			if month == currentMonth {
				result.currentMeterReadings = append(result.currentMeterReadings, todo)
			} else {
				result.meterReadings = append(result.meterReadings, todo)
			}
		}
		if generationCount > 0 {
			result.billGenerations = append(result.billGenerations, *newHomeTodo(
				unit.Name, fmt.Sprintf("%s 月还有 %d 间房待生成账单", month, generationCount), todoTagBillGeneration,
			))
		}
	}

	unpaidPeriods := make([]string, 0, len(unpaidCountByPeriod))
	for month := range unpaidCountByPeriod {
		unpaidPeriods = append(unpaidPeriods, month)
	}
	sort.Strings(unpaidPeriods)
	for _, month := range unpaidPeriods {
		result.rentCollections = append(result.rentCollections, *newHomeTodo(
			unit.Name, fmt.Sprintf("%s 月还有 %d 笔租金待收", month, unpaidCountByPeriod[month]), todoTagRentCollection,
		))
	}
	return result, nil
}

func (s *Service) listUnitReadingsByPeriod(
	ctx context.Context,
	unit *domainunit.Unit,
	currentMonth string,
) (map[string]map[int]map[int]*domainmeterreading.MeterReading, error) {
	readingsByPeriod := make(map[string]map[int]map[int]*domainmeterreading.MeterReading)
	if s.meterReadingRepo == nil {
		return readingsByPeriod, nil
	}
	readings, err := s.meterReadingRepo.List(ctx, unit.AccountID, domainmeterreading.ListQuery{
		UnitID: unit.ID, PeriodMonthTo: currentMonth,
	})
	if err != nil {
		return nil, err
	}
	for _, reading := range readings {
		if reading == nil {
			continue
		}
		if readingsByPeriod[reading.PeriodMonth] == nil {
			readingsByPeriod[reading.PeriodMonth] = make(map[int]map[int]*domainmeterreading.MeterReading)
		}
		if readingsByPeriod[reading.PeriodMonth][reading.RoomID] == nil {
			readingsByPeriod[reading.PeriodMonth][reading.RoomID] = make(map[int]*domainmeterreading.MeterReading)
		}
		readingsByPeriod[reading.PeriodMonth][reading.RoomID][reading.MeterItemID] = reading
	}
	return readingsByPeriod, nil
}

func (s *Service) canViewUnit(
	ctx context.Context,
	accountID string,
	unitID int,
	permissionCode string,
) (bool, error) {
	if s.rbacService == nil {
		return true, nil
	}
	if err := s.rbacService.RequireUnitPermission(
		ctx,
		accountID,
		unitID,
		permissionCode,
	); err != nil {
		if errors.Is(err, domainrbac.ErrForbidden) ||
			errors.Is(err, domainrbac.ErrUnitNotInAccount) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isRoomMeterReadingComplete(
	room *domainroom.Room,
	readings map[int]*domainmeterreading.MeterReading,
) (bool, bool) {
	meterItemIDs := roomMeterItemIDs(room)
	if len(meterItemIDs) == 0 {
		return true, false
	}
	for _, meterItemID := range meterItemIDs {
		reading := readings[meterItemID]
		if reading == nil || reading.UsageAmount == nil {
			return false, true
		}
	}
	return true, true
}

func roomMeterItemIDs(room *domainroom.Room) []int {
	if room == nil {
		return nil
	}
	ids := make([]int, 0, len(room.MeterItems))
	for _, item := range room.MeterItems {
		if item == nil {
			continue
		}
		ids = append(ids, item.ID)
	}
	if len(ids) > 0 {
		return ids
	}
	return room.MeterItemIDs
}

func newHomeTodo(title string, description string, tag string) *Todo {
	return &Todo{
		Title:       title,
		Description: description,
		Tag:         tag,
	}
}

type unitTodoResult struct {
	currentMeterReadings []Todo
	meterReadings        []Todo
	billGenerations      []Todo
	rentCollections      []Todo
}

type homeTodoBuckets struct {
	currentMeterReadings []Todo
	meterReadings        []Todo
	billGenerations      []Todo
	rentCollections      []Todo
}

func (b *homeTodoBuckets) append(result *unitTodoResult) {
	if result == nil {
		return
	}
	b.currentMeterReadings = append(b.currentMeterReadings, result.currentMeterReadings...)
	b.meterReadings = append(b.meterReadings, result.meterReadings...)
	b.billGenerations = append(b.billGenerations, result.billGenerations...)
	b.rentCollections = append(b.rentCollections, result.rentCollections...)
}

func (b *homeTodoBuckets) todos(limit int) []Todo {
	todos := make([]Todo, 0, len(b.currentMeterReadings)+len(b.meterReadings)+len(b.billGenerations)+len(b.rentCollections))
	todos = append(todos, b.currentMeterReadings...)
	todos = append(todos, b.meterReadings...)
	todos = append(todos, b.billGenerations...)
	todos = append(todos, b.rentCollections...)
	if limit > 0 && len(todos) > limit {
		return todos[:limit]
	}
	return todos
}

// currentAndPreviousMonth 返回当前月份和上一月份的 YYYY-MM 表示。
func (s *Service) currentAndPreviousMonth() (string, string) {
	return currentAndPreviousMonth(s.currentTime())
}

func currentAndPreviousMonth(now time.Time) (string, string) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return monthStart.Format("2006-01"), monthStart.AddDate(0, -1, 0).Format("2006-01")
}

func (s *Service) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func meterReadingTodoDescription(unit *domainunit.Unit, month string, missingCount int, now time.Time) string {
	now = now.In(homeLocation)
	if month < now.Format("2006-01") {
		return fmt.Sprintf("%s 月抄表已逾期，还有 %d 间房未抄表", month, missingCount)
	}
	meterReadingDay := unit.Config.MeterReadingDay
	if meterReadingDay < domainunit.MinMeterReadingDay || meterReadingDay > domainunit.MaxMeterReadingDay {
		meterReadingDay = domainunit.DefaultMeterReadingDay
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, homeLocation)
	dueDate := meterReadingDueDate(monthStart, meterReadingDay)
	switch {
	case now.Day() < dueDate.Day():
		return fmt.Sprintf("%s 月还有 %d 间房待抄表，计划 %s 抄表", month, missingCount, dueDate.Format("2006-01-02"))
	case now.Day() == dueDate.Day():
		return fmt.Sprintf("今日需完成 %s 月抄表，还有 %d 间房未抄表", month, missingCount)
	default:
		return fmt.Sprintf("%s 月抄表已逾期，还有 %d 间房未抄表", month, missingCount)
	}
}

func meterReadingDueDate(monthStart time.Time, meterReadingDay int) time.Time {
	lastDay := monthStart.AddDate(0, 1, -1).Day()
	if meterReadingDay > lastDay {
		meterReadingDay = lastDay
	}
	return time.Date(
		monthStart.Year(),
		monthStart.Month(),
		meterReadingDay,
		0,
		0,
		0,
		0,
		monthStart.Location(),
	)
}
