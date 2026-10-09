package home

import (
	"context"
	"errors"
	"fmt"
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
func NewService(unitRepo domainunit.Repository, roomRepo domainroom.Repository, rentBillRepo domainrentbill.Repository, meterReadingRepo domainmeterreading.Repository, rbacRepo domainrbac.Repository) *Service {
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

	now := s.currentTime()
	currentMonth, previousMonth := currentAndPreviousMonth(now)
	summary := &Summary{CurrentMonth: currentMonth}
	todoBuckets := &homeTodoBuckets{}

	for _, unit := range units {
		if unit == nil {
			continue
		}
		summary.BuildingCount++

		// 查询出所有已出租的房间
		rentedRooms, err := s.listRoomsInUnit(ctx, unit, new(false))
		if err != nil {
			return nil, err
		}

		// 统计空房的数量
		availableCount, err := s.countRoomsInUnit(ctx, unit, new(true))
		if err != nil {
			return nil, err
		}

		summary.RentedRoomCount += len(rentedRooms)
		summary.VacantRoomCount += availableCount
		summary.RoomCount += len(rentedRooms) + availableCount
		err = s.collectRentBillSummary(ctx, unit, summary, currentMonth, previousMonth)
		if err != nil {
			return nil, err
		}
		todoPeriodMonth := unitTodoPeriodMonth(unit, now)
		unitTodos, err := s.collectUnitTodos(ctx, accountID, unit, rentedRooms, todoPeriodMonth)
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
func (s *Service) listRoomsInUnit(ctx context.Context, unit *domainunit.Unit, isActive *bool) ([]*domainroom.Room, error) {
	return s.roomRepo.List(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: isActive,
	})
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

func (s *Service) collectUnitTodos(ctx context.Context, accountID string, unit *domainunit.Unit, rentedRooms []*domainroom.Room, periodMonth string) (*unitTodoResult, error) {
	result := &unitTodoResult{}
	// 首页待办只针对已出租房间；可用空房不需要抄表、出账或收租。
	if len(rentedRooms) == 0 {
		return result, nil
	}

	// 先按当前账号权限判断哪些类型的待办允许展示。
	canViewMeterReading, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionMeterReadingView)
	if err != nil {
		return nil, err
	}
	canViewRentBill, err := s.canViewUnit(ctx, accountID, unit.ID, domainrbac.PermissionRentBillView)
	if err != nil {
		return nil, err
	}

	// 抄表待办和账单生成都依赖目标月份的抄表完成度；无权限时跳过查询。
	needReadings := canViewMeterReading || canViewRentBill
	readingsByRoom, err := s.listUnitReadingsByRoom(ctx, unit, periodMonth, needReadings)
	if err != nil {
		return nil, err
	}

	// 逐房间记录抄表是否完成，并统计有抄表项但未完整抄表的房间。
	completeByRoomID := make(map[int]bool, len(rentedRooms))
	missingMeterRoomCount := 0
	for _, room := range rentedRooms {
		if room == nil {
			continue
		}
		complete, hasMeterItems := isRoomMeterReadingComplete(room, readingsByRoom[room.ID])
		completeByRoomID[room.ID] = complete
		if hasMeterItems && !complete {
			missingMeterRoomCount++
		}
	}
	if canViewMeterReading && missingMeterRoomCount > 0 {
		result.meterReading = newHomeTodo(unit.Name, fmt.Sprintf("还有 %d 间房未抄表", missingMeterRoomCount), todoTagMeterReading)
	}

	// 没有账单查看权限时，只返回前面已经收集到的抄表待办。
	if !canViewRentBill || s.rentBillRepo == nil {
		return result, nil
	}

	// 查询楼栋目标账单月份已有账单，用于判断哪些房间已出账、哪些账单还待收。
	bills, err := s.rentBillRepo.List(ctx, unit.AccountID, domainrentbill.ListQuery{
		UnitID:      unit.ID,
		PeriodMonth: periodMonth,
	})
	if err != nil {
		return nil, err
	}
	billRoomIDs := make(map[int]struct{}, len(bills))
	unpaidBillCount := 0
	for _, bill := range bills {
		if bill == nil {
			continue
		}
		billRoomIDs[bill.RoomID] = struct{}{}
		if bill.Status == domainrentbill.StatusUnpaid {
			unpaidBillCount++
		}
	}

	// 只有抄表完成且目标月份还没有账单的已租房间，才进入待生成账单统计。
	pendingBillGenerationRoomCount := 0
	for _, room := range rentedRooms {
		if room == nil || !completeByRoomID[room.ID] {
			continue
		}
		if _, ok := billRoomIDs[room.ID]; !ok {
			pendingBillGenerationRoomCount++
		}
	}
	if pendingBillGenerationRoomCount > 0 {
		result.billGeneration = newHomeTodo(unit.Name, fmt.Sprintf("还有 %d 间房待生成账单", pendingBillGenerationRoomCount), todoTagBillGeneration)
	}
	if unpaidBillCount > 0 {
		result.rentCollection = newHomeTodo(unit.Name, fmt.Sprintf("还有 %d 笔租金待收", unpaidBillCount), todoTagRentCollection)
	}
	return result, nil
}

func (s *Service) listUnitReadingsByRoom(ctx context.Context, unit *domainunit.Unit, periodMonth string, enabled bool) (map[int]map[int]*domainmeterreading.MeterReading, error) {
	readingsByRoom := make(map[int]map[int]*domainmeterreading.MeterReading)
	if !enabled || s.meterReadingRepo == nil {
		return readingsByRoom, nil
	}
	readings, err := s.meterReadingRepo.List(ctx, unit.AccountID, domainmeterreading.ListQuery{
		UnitID:      unit.ID,
		PeriodMonth: periodMonth,
	})
	if err != nil {
		return nil, err
	}
	for _, reading := range readings {
		if reading == nil {
			continue
		}
		readingsByMeterItem := readingsByRoom[reading.RoomID]
		if readingsByMeterItem == nil {
			readingsByMeterItem = make(map[int]*domainmeterreading.MeterReading)
			readingsByRoom[reading.RoomID] = readingsByMeterItem
		}
		readingsByMeterItem[reading.MeterItemID] = reading
	}
	return readingsByRoom, nil
}

func (s *Service) canViewUnit(ctx context.Context, accountID string, unitID int, permissionCode string) (bool, error) {
	if s.rbacService == nil {
		return true, nil
	}
	if err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		if errors.Is(err, domainrbac.ErrForbidden) || errors.Is(err, domainrbac.ErrUnitNotInAccount) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isRoomMeterReadingComplete(room *domainroom.Room, readings map[int]*domainmeterreading.MeterReading) (bool, bool) {
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
	meterReading   *Todo
	billGeneration *Todo
	rentCollection *Todo
}

type homeTodoBuckets struct {
	meterReadings   []Todo
	billGenerations []Todo
	rentCollections []Todo
}

func (b *homeTodoBuckets) append(result *unitTodoResult) {
	if result == nil {
		return
	}
	if result.meterReading != nil {
		b.meterReadings = append(b.meterReadings, *result.meterReading)
	}
	if result.billGeneration != nil {
		b.billGenerations = append(b.billGenerations, *result.billGeneration)
	}
	if result.rentCollection != nil {
		b.rentCollections = append(b.rentCollections, *result.rentCollection)
	}
}

func (b *homeTodoBuckets) todos(limit int) []Todo {
	todos := make([]Todo, 0, len(b.meterReadings)+len(b.billGenerations)+len(b.rentCollections))
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

func unitTodoPeriodMonth(unit *domainunit.Unit, now time.Time) string {
	meterReadingDay := domainunit.DefaultMeterReadingDay
	if unit != nil && unit.Config.MeterReadingDay != 0 {
		meterReadingDay = unit.Config.MeterReadingDay
	}
	return todoPeriodMonthByMeterReadingDay(now, meterReadingDay)
}

func (s *Service) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func todoPeriodMonthByMeterReadingDay(now time.Time, meterReadingDay int) string {
	if meterReadingDay < domainunit.MinMeterReadingDay || meterReadingDay > domainunit.MaxMeterReadingDay {
		meterReadingDay = domainunit.DefaultMeterReadingDay
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	currentDueDate := meterReadingDueDate(monthStart, meterReadingDay)

	offsetMonths := -2
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if !today.Before(currentDueDate) {
		offsetMonths = -1
	}
	return monthStart.AddDate(0, offsetMonths, 0).Format("2006-01")
}

func meterReadingDueDate(monthStart time.Time, meterReadingDay int) time.Time {
	lastDay := monthStart.AddDate(0, 1, -1).Day()
	if meterReadingDay > lastDay {
		meterReadingDay = lastDay
	}
	return time.Date(monthStart.Year(), monthStart.Month(), meterReadingDay, 0, 0, 0, 0, monthStart.Location())
}
