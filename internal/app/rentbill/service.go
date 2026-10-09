package rentbill

import (
	"context"
	"errors"
	"strings"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

// Service 提供租金账单用例。
type Service struct {
	repo             domainrentbill.Repository
	roomRepo         domainroom.Repository
	unitRepo         domainunit.Repository
	meterReadingRepo domainmeterreading.Repository
	rbacService      domainrbac.IRbacService
}

// NewService 创建租金账单服务。
func NewService(repo domainrentbill.Repository, roomRepo domainroom.Repository, unitRepo domainunit.Repository, meterReadingRepo domainmeterreading.Repository, rbacRepo domainrbac.Repository) *Service {
	return &Service{
		repo:             repo,
		roomRepo:         roomRepo,
		unitRepo:         unitRepo,
		meterReadingRepo: meterReadingRepo,
		rbacService:      domainrbac.NewRbacService(rbacRepo),
	}
}

// ListRoomRentBills 获取房间账单列表。
func (s *Service) ListRoomRentBills(ctx context.Context, query *RoomRentBillsQuery) ([]*domainrentbill.RentBill, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	room, unit, err := s.requireRoomPermission(ctx, accountID, query.RoomID, domainrbac.PermissionRentBillView)
	if err != nil {
		return nil, err
	}
	periodMonth := strings.TrimSpace(query.PeriodMonth)
	if periodMonth != "" {
		var err error
		periodMonth, err = validatePeriodMonth(periodMonth)
		if err != nil {
			return nil, err
		}
	}
	return s.repo.List(ctx, unit.AccountID, domainrentbill.ListQuery{
		RoomID:      room.ID,
		PeriodMonth: periodMonth,
		Status:      query.Status,
		Limit:       query.Limit,
		Offset:      query.Offset,
	})
}

// ListUnitRentBills 获取单元账单列表。
func (s *Service) ListUnitRentBills(ctx context.Context, query *UnitRentBillsQuery) ([]*domainrentbill.RentBill, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	unit, err := s.requireUnitPermission(ctx, accountID, query.UnitID, domainrbac.PermissionRentBillView)
	if err != nil {
		return nil, err
	}
	periodMonth := strings.TrimSpace(query.PeriodMonth)
	if periodMonth != "" {
		var err error
		periodMonth, err = validatePeriodMonth(periodMonth)
		if err != nil {
			return nil, err
		}
	}
	return s.repo.List(ctx, unit.AccountID, domainrentbill.ListQuery{
		UnitID:      unit.ID,
		PeriodMonth: periodMonth,
		Status:      query.Status,
		Limit:       query.Limit,
		Offset:      query.Offset,
	})
}

// GenerateRoomRentBill 生成或刷新房间月度账单。
func (s *Service) GenerateRoomRentBill(ctx context.Context, cmd *GenerateRoomRentBillCommand) (*GeneratedRentBills, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	periodMonth, err := validatePeriodMonth(cmd.PeriodMonth)
	if err != nil {
		return nil, err
	}
	room, unit, err := s.requireRoomPermission(ctx, accountID, cmd.RoomID, domainrbac.PermissionRentBillManage)
	if err != nil {
		return nil, err
	}
	if existing, skipped, err := s.skippableExistingBill(ctx, unit.AccountID, room.ID, periodMonth); err != nil {
		return nil, err
	} else if skipped {
		return &GeneratedRentBills{Skipped: []*domainrentbill.RentBill{existing}}, nil
	}
	bill, err := s.buildGeneratedBill(ctx, unit.AccountID, room, periodMonth, cmd.Remark)
	if err != nil {
		return nil, err
	}
	return s.saveGeneratedBills(ctx, unit.AccountID, []*domainrentbill.RentBill{bill})
}

// GenerateUnitRentBills 批量生成或刷新单元月度账单。
func (s *Service) GenerateUnitRentBills(ctx context.Context, cmd *GenerateUnitRentBillsCommand) (*GeneratedRentBills, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	periodMonth, err := validatePeriodMonth(cmd.PeriodMonth)
	if err != nil {
		return nil, err
	}
	unit, err := s.requireUnitPermission(ctx, accountID, cmd.UnitID, domainrbac.PermissionRentBillManage)
	if err != nil {
		return nil, err
	}
	rented := false
	rooms, err := s.roomRepo.List(ctx, unit.AccountID, domainroom.ListQuery{
		UnitID:   unit.ID,
		IsActive: &rented,
	})
	if err != nil {
		return nil, err
	}
	bills := make([]*domainrentbill.RentBill, 0, len(rooms))
	skippedBills := make([]*domainrentbill.RentBill, 0)
	for _, room := range rooms {
		if room == nil {
			continue
		}
		if existing, skipped, err := s.skippableExistingBill(ctx, unit.AccountID, room.ID, periodMonth); err != nil {
			return nil, err
		} else if skipped {
			skippedBills = append(skippedBills, existing)
			continue
		}
		bill, err := s.buildGeneratedBill(ctx, unit.AccountID, room, periodMonth, cmd.Remark)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	result, err := s.saveGeneratedBills(ctx, unit.AccountID, bills)
	if err != nil {
		return nil, err
	}
	result.Skipped = append(skippedBills, result.Skipped...)
	return result, nil
}

// GetRentBill 获取账单详情。
func (s *Service) GetRentBill(ctx context.Context, id int) (*domainrentbill.RentBill, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bill, unit, err := s.requireBillPermission(ctx, accountID, id, domainrbac.PermissionRentBillView)
	if err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, unit.AccountID, bill.ID)
}

// MarkRentBillPaid 标记账单已收款。
func (s *Service) MarkRentBillPaid(ctx context.Context, cmd *MarkRentBillPaidCommand) (*domainrentbill.RentBill, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bill, unit, err := s.requireBillPermission(ctx, accountID, cmd.RentBillID, domainrbac.PermissionRentBillManage)
	if err != nil {
		return nil, err
	}
	paidAt := time.Now()
	if cmd.PaidAt != nil {
		paidAt = *cmd.PaidAt
	}
	next := *bill
	if err := next.MarkPaid(paidAt); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, unit.AccountID, &next)
}

// CancelRentBill 取消未支付账单。
func (s *Service) CancelRentBill(ctx context.Context, cmd *CancelRentBillCommand) (*domainrentbill.RentBill, error) {
	accountID, err := appcore.AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bill, unit, err := s.requireBillPermission(ctx, accountID, cmd.RentBillID, domainrbac.PermissionRentBillManage)
	if err != nil {
		return nil, err
	}
	next := *bill
	if err := next.Cancel(); err != nil {
		return nil, err
	}
	if cmd.Remark != nil {
		next.Remark = normalizeRemark(cmd.Remark)
	}
	return s.repo.Update(ctx, unit.AccountID, &next)
}

func (s *Service) buildGeneratedBill(ctx context.Context, accountID int, room *domainroom.Room, periodMonth string, remark *string) (*domainrentbill.RentBill, error) {
	meterFee, err := s.calculateRoomMeterFee(ctx, accountID, room, periodMonth)
	if err != nil {
		return nil, err
	}
	return domainrentbill.GenerateRentBill(domainrentbill.GenerateParams{
		RoomID:      room.ID,
		PeriodMonth: periodMonth,
		BaseRent:    room.Price,
		MeterFee:    meterFee,
		Remark:      remark,
	})
}

func (s *Service) calculateRoomMeterFee(ctx context.Context, accountID int, room *domainroom.Room, periodMonth string) (int, error) {
	if len(room.MeterItems) == 0 {
		return 0, nil
	}
	readings, err := s.meterReadingRepo.List(ctx, accountID, domainmeterreading.ListQuery{
		RoomID:      room.ID,
		PeriodMonth: periodMonth,
	})
	if err != nil {
		return 0, err
	}
	readingByMeterItemID := make(map[int]*domainmeterreading.MeterReading, len(readings))
	for _, reading := range readings {
		if reading == nil {
			continue
		}
		readingByMeterItemID[reading.MeterItemID] = reading
	}
	total := 0
	for _, item := range room.MeterItems {
		if item == nil {
			continue
		}
		reading := readingByMeterItemID[item.ID]
		if reading == nil || reading.UsageAmount == nil {
			return 0, domainrentbill.ErrMeterReadingMissing
		}
		fee, err := reading.CalculateFee(item.UnitPrice)
		if err != nil {
			return 0, err
		}
		total += fee
	}
	return total, nil
}

func (s *Service) saveGeneratedBills(ctx context.Context, accountID int, bills []*domainrentbill.RentBill) (*GeneratedRentBills, error) {
	if len(bills) == 0 {
		return &GeneratedRentBills{}, nil
	}
	result, err := s.repo.SaveGeneratedBills(ctx, accountID, bills)
	if err != nil {
		return nil, err
	}
	return &GeneratedRentBills{
		Saved:   result.Saved,
		Skipped: result.Skipped,
	}, nil
}

func (s *Service) skippableExistingBill(ctx context.Context, accountID int, roomID int, periodMonth string) (*domainrentbill.RentBill, bool, error) {
	bill, err := s.repo.GetByRoomPeriod(ctx, accountID, roomID, periodMonth)
	if err != nil {
		if errors.Is(err, domainrentbill.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return bill, bill.Status != domainrentbill.StatusUnpaid, nil
}

func (s *Service) requireBillPermission(ctx context.Context, accountID int, billID int, permissionCode string) (*domainrentbill.RentBill, *domainunit.Unit, error) {
	bill, err := s.repo.GetByID(ctx, billID)
	if err != nil {
		return nil, nil, err
	}
	room, unit, err := s.requireRoomPermission(ctx, accountID, bill.RoomID, permissionCode)
	if err != nil {
		return nil, nil, err
	}
	bill.RoomID = room.ID
	return bill, unit, nil
}

func (s *Service) requireRoomPermission(ctx context.Context, accountID int, roomID int, permissionCode string) (*domainroom.Room, *domainunit.Unit, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, nil, err
	}
	unit, err := s.requireUnitPermission(ctx, accountID, room.UnitID, permissionCode)
	if err != nil {
		return nil, nil, err
	}
	return room, unit, nil
}

func (s *Service) requireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (*domainunit.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}
	if unit.AccountID == accountID {
		return unit, nil
	}
	if err := s.rbacService.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		return nil, err
	}
	return unit, nil
}

func validatePeriodMonth(periodMonth string) (string, error) {
	periodMonth = strings.TrimSpace(periodMonth)
	if periodMonth == "" {
		return "", domainrentbill.ErrPeriodMonthRequired
	}
	if _, err := time.Parse("2006-01", periodMonth); err != nil {
		return "", domainrentbill.ErrInvalidPeriodMonth
	}
	return periodMonth, nil
}

func normalizeRemark(remark *string) *string {
	if remark == nil {
		return nil
	}
	value := strings.TrimSpace(*remark)
	if value == "" {
		return nil
	}
	return &value
}
