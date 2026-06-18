package rentbill

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
	domainmeterreading "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meterreading"
	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	domainrentbill "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rentbill"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
)

func TestGenerateRoomRentBillUsesRoomRentAndMeterFee(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200}
	power := &domainmeteritem.MeterItem{ID: 2, Code: "电", Unit: "kWh", UnitPrice: 150}
	room := &domainroom.Room{ID: 11, UnitID: 7, Price: 100000, MeterItems: []*domainmeteritem.MeterItem{water, power}}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	waterUsage := 3.0
	powerUsage := 4.0
	rentBillRepo := &fakeRentBillRepo{}
	service := newTestService(rentBillRepo, []*domainroom.Room{room}, []*domainunit.Unit{unit}, map[string][]*domainmeterreading.MeterReading{
		meterReadingKey(11, "2026-06"): {
			{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-06", UsageAmount: &waterUsage},
			{RoomID: 11, MeterItemID: 2, PeriodMonth: "2026-06", UsageAmount: &powerUsage},
		},
	}, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)

	result, err := service.GenerateRoomRentBill(ctx, &GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatalf("GenerateRoomRentBill() error = %v", err)
	}
	if len(result.Saved) != 1 {
		t.Fatalf("saved count = %d, want 1", len(result.Saved))
	}
	bill := result.Saved[0]
	if bill.BaseRent != 100000 {
		t.Fatalf("base rent = %d, want 100000", bill.BaseRent)
	}
	if bill.MeterFee != 1200 {
		t.Fatalf("meter fee = %d, want 1200", bill.MeterFee)
	}
	if bill.TotalAmount != 101200 {
		t.Fatalf("total amount = %d, want 101200", bill.TotalAmount)
	}
	if len(rentBillRepo.saved) != 1 {
		t.Fatalf("repo saved count = %d, want 1", len(rentBillRepo.saved))
	}
}

func TestGenerateRoomRentBillRejectsMissingMeterReading(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200}
	room := &domainroom.Room{ID: 11, UnitID: 7, Price: 100000, MeterItems: []*domainmeteritem.MeterItem{water}}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	rentBillRepo := &fakeRentBillRepo{}
	service := newTestService(rentBillRepo, []*domainroom.Room{room}, []*domainunit.Unit{unit}, nil, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)

	_, err := service.GenerateRoomRentBill(ctx, &GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
	if !errors.Is(err, domainrentbill.ErrMeterReadingMissing) {
		t.Fatalf("error = %v, want ErrMeterReadingMissing", err)
	}
	if len(rentBillRepo.saved) != 0 {
		t.Fatalf("repo saved count = %d, want 0", len(rentBillRepo.saved))
	}
}

func TestGenerateUnitRentBillsSkipsPaidExistingBill(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200}
	room := &domainroom.Room{ID: 11, UnitID: 7, Price: 100000, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	paidAt := time.Now()
	existing := &domainrentbill.RentBill{ID: 99, RoomID: 11, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, PaidAt: &paidAt}
	rentBillRepo := &fakeRentBillRepo{byRoomPeriod: map[string]*domainrentbill.RentBill{rentBillKey(11, "2026-06"): existing}}
	service := newTestService(rentBillRepo, []*domainroom.Room{room}, []*domainunit.Unit{unit}, nil, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)

	result, err := service.GenerateUnitRentBills(ctx, &GenerateUnitRentBillsCommand{UnitID: 7, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatalf("GenerateUnitRentBills() error = %v", err)
	}
	if len(result.Saved) != 0 {
		t.Fatalf("saved count = %d, want 0", len(result.Saved))
	}
	if len(result.Skipped) != 1 || result.Skipped[0].ID != 99 {
		t.Fatalf("skipped = %#v, want bill 99", result.Skipped)
	}
	if len(rentBillRepo.saved) != 0 {
		t.Fatalf("repo saved count = %d, want 0", len(rentBillRepo.saved))
	}
}

func TestGenerateUnitRentBillsSkipsAvailableRooms(t *testing.T) {
	room := &domainroom.Room{ID: 11, UnitID: 7, Price: 100000, IsActive: true}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	rentBillRepo := &fakeRentBillRepo{}
	service := newTestService(rentBillRepo, []*domainroom.Room{room}, []*domainunit.Unit{unit}, nil, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)

	result, err := service.GenerateUnitRentBills(ctx, &GenerateUnitRentBillsCommand{UnitID: 7, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatalf("GenerateUnitRentBills() error = %v", err)
	}
	if len(result.Saved) != 0 {
		t.Fatalf("saved count = %d, want 0", len(result.Saved))
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("skipped count = %d, want 0", len(result.Skipped))
	}
	if len(rentBillRepo.saved) != 0 {
		t.Fatalf("repo saved count = %d, want 0", len(rentBillRepo.saved))
	}
}

func TestGenerateRoomRentBillChecksGrantedPermission(t *testing.T) {
	room := &domainroom.Room{ID: 11, UnitID: 7, Price: 100000}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	rbac := &fakeRbacService{}
	service := newTestService(&fakeRentBillRepo{}, []*domainroom.Room{room}, []*domainunit.Unit{unit}, nil, rbac)
	ctx := appcore.SetAccountIDToContext(context.Background(), 30)

	_, err := service.GenerateRoomRentBill(ctx, &GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatalf("GenerateRoomRentBill() error = %v", err)
	}
	if rbac.permissionCode != domainrbac.PermissionRentBillManage {
		t.Fatalf("permission = %q, want %q", rbac.permissionCode, domainrbac.PermissionRentBillManage)
	}
	if rbac.accountID != 30 || rbac.unitID != 7 {
		t.Fatalf("rbac got account=%d unit=%d, want account=30 unit=7", rbac.accountID, rbac.unitID)
	}
}

func newTestService(rentBillRepo *fakeRentBillRepo, rooms []*domainroom.Room, units []*domainunit.Unit, readings map[string][]*domainmeterreading.MeterReading, rbac *fakeRbacService) *Service {
	roomByID := make(map[int]*domainroom.Room, len(rooms))
	for _, room := range rooms {
		roomByID[room.ID] = room
	}
	unitByID := make(map[int]*domainunit.Unit, len(units))
	for _, unit := range units {
		unitByID[unit.ID] = unit
	}
	return &Service{
		repo:             rentBillRepo,
		roomRepo:         &fakeRoomRepo{byID: roomByID},
		unitRepo:         &fakeUnitRepo{byID: unitByID},
		meterReadingRepo: &fakeMeterReadingRepo{listByRoomPeriod: readings},
		rbacService:      rbac,
	}
}

type fakeRentBillRepo struct {
	byID         map[int]*domainrentbill.RentBill
	byRoomPeriod map[string]*domainrentbill.RentBill
	saved        []*domainrentbill.RentBill
	updated      []*domainrentbill.RentBill
}

func (r *fakeRentBillRepo) Create(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	return bill, nil
}

func (r *fakeRentBillRepo) Get(ctx context.Context, accountID int, id int) (*domainrentbill.RentBill, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeRentBillRepo) GetByID(ctx context.Context, id int) (*domainrentbill.RentBill, error) {
	if r.byID != nil && r.byID[id] != nil {
		return r.byID[id], nil
	}
	return nil, domainrentbill.ErrNotFound
}

func (r *fakeRentBillRepo) GetByRoomPeriod(ctx context.Context, accountID int, roomID int, periodMonth string) (*domainrentbill.RentBill, error) {
	if r.byRoomPeriod != nil && r.byRoomPeriod[rentBillKey(roomID, periodMonth)] != nil {
		return r.byRoomPeriod[rentBillKey(roomID, periodMonth)], nil
	}
	return nil, domainrentbill.ErrNotFound
}

func (r *fakeRentBillRepo) List(ctx context.Context, accountID int, query domainrentbill.ListQuery) ([]*domainrentbill.RentBill, error) {
	return nil, nil
}

func (r *fakeRentBillRepo) SaveGeneratedBills(ctx context.Context, accountID int, bills []*domainrentbill.RentBill) (*domainrentbill.SaveGeneratedBillsResult, error) {
	r.saved = append([]*domainrentbill.RentBill(nil), bills...)
	result := &domainrentbill.SaveGeneratedBillsResult{Saved: make([]*domainrentbill.RentBill, 0, len(bills))}
	for index, bill := range bills {
		if bill.ID == 0 {
			bill.ID = index + 1
		}
		result.Saved = append(result.Saved, bill)
	}
	return result, nil
}

func (r *fakeRentBillRepo) Update(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	r.updated = append(r.updated, bill)
	return bill, nil
}

func (r *fakeRentBillRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeMeterReadingRepo struct {
	listByRoomPeriod map[string][]*domainmeterreading.MeterReading
}

func (r *fakeMeterReadingRepo) Create(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	return reading, nil
}

func (r *fakeMeterReadingRepo) Get(ctx context.Context, accountID int, id int) (*domainmeterreading.MeterReading, error) {
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeMeterReadingRepo) GetByRoomMeterItemPeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeMeterReadingRepo) GetLatestBeforePeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeMeterReadingRepo) List(ctx context.Context, accountID int, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	return r.listByRoomPeriod[meterReadingKey(query.RoomID, query.PeriodMonth)], nil
}

func (r *fakeMeterReadingRepo) SaveRoomMonth(ctx context.Context, accountID int, readings []*domainmeterreading.MeterReading) ([]*domainmeterreading.MeterReading, error) {
	return readings, nil
}

func (r *fakeMeterReadingRepo) Update(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	return reading, nil
}

func (r *fakeMeterReadingRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeRoomRepo struct {
	byID map[int]*domainroom.Room
}

func (r *fakeRoomRepo) Create(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	return room, nil
}

func (r *fakeRoomRepo) Get(ctx context.Context, accountID int, id int) (*domainroom.Room, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeRoomRepo) GetByID(ctx context.Context, id int) (*domainroom.Room, error) {
	room := r.byID[id]
	if room == nil {
		return nil, domainroom.ErrNotFound
	}
	return room, nil
}

func (r *fakeRoomRepo) List(ctx context.Context, accountID int, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	rooms := make([]*domainroom.Room, 0, len(r.byID))
	for _, room := range r.byID {
		if room.UnitID != query.UnitID {
			continue
		}
		if query.IsActive != nil && room.IsActive != *query.IsActive {
			continue
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *fakeRoomRepo) Count(ctx context.Context, accountID int, query domainroom.ListQuery) (int, error) {
	rooms, err := r.List(ctx, accountID, query)
	return len(rooms), err
}

func (r *fakeRoomRepo) Update(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	return room, nil
}

func (r *fakeRoomRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeUnitRepo struct {
	byID map[int]*domainunit.Unit
}

func (r *fakeUnitRepo) Create(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeUnitRepo) Get(ctx context.Context, accountID int, id int) (*domainunit.Unit, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeUnitRepo) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	unit := r.byID[id]
	if unit == nil {
		return nil, domainunit.ErrNotFound
	}
	return unit, nil
}

func (r *fakeUnitRepo) List(ctx context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	return nil, nil
}

func (r *fakeUnitRepo) Update(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeUnitRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeRbacService struct {
	accountID      int
	unitID         int
	permissionCode string
	err            error
}

func (s *fakeRbacService) RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error {
	s.accountID = accountID
	s.unitID = unitID
	s.permissionCode = permissionCode
	return s.err
}

func rentBillKey(roomID int, periodMonth string) string {
	return fmt.Sprintf("%d:%s", roomID, periodMonth)
}

func meterReadingKey(roomID int, periodMonth string) string {
	return fmt.Sprintf("%d:%s", roomID, periodMonth)
}

var _ domainrentbill.Repository = (*fakeRentBillRepo)(nil)
var _ domainmeterreading.Repository = (*fakeMeterReadingRepo)(nil)
var _ domainroom.Repository = (*fakeRoomRepo)(nil)
var _ domainunit.Repository = (*fakeUnitRepo)(nil)
var _ domainrbac.IRbacService = (*fakeRbacService)(nil)
