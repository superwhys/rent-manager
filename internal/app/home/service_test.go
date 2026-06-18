package home

import (
	"context"
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

func TestGetHomeSummaryBuildsDynamicTodos(t *testing.T) {
	periodMonth := time.Now().Format("2006-01")
	unit := &domainunit.Unit{ID: 1, AccountID: 10, Name: "幸福小区 · 3栋"}
	water := &domainmeteritem.MeterItem{ID: 1}
	rooms := []*domainroom.Room{
		{ID: 11, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 12, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 13, UnitID: 1, IsActive: false},
		{ID: 14, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 15, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 16, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 17, UnitID: 1, IsActive: true, MeterItems: []*domainmeteritem.MeterItem{water}},
	}
	usage := 1.0
	paidAt := time.Now()
	service := newHomeTestService(
		[]*domainunit.Unit{unit},
		rooms,
		[]*domainmeterreading.MeterReading{
			{RoomID: 12, MeterItemID: 1, PeriodMonth: periodMonth, UsageAmount: &usage},
			{RoomID: 14, MeterItemID: 1, PeriodMonth: periodMonth, UsageAmount: &usage},
			{RoomID: 15, MeterItemID: 1, PeriodMonth: periodMonth, UsageAmount: &usage},
			{RoomID: 16, MeterItemID: 1, PeriodMonth: periodMonth, UsageAmount: &usage},
		},
		[]*domainrentbill.RentBill{
			{ID: 1, RoomID: 14, PeriodMonth: periodMonth, Status: domainrentbill.StatusUnpaid, TotalAmount: 1000},
			{ID: 2, RoomID: 15, PeriodMonth: periodMonth, Status: domainrentbill.StatusPaid, PaidAt: &paidAt, TotalAmount: 1000},
			{ID: 3, RoomID: 16, PeriodMonth: periodMonth, Status: domainrentbill.StatusCancelled, TotalAmount: 1000},
		},
		&fakeHomeRbacService{},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	summary, err := service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatalf("GetHomeSummary() error = %v", err)
	}
	if summary.RentedRoomCount != 6 {
		t.Fatalf("rented room count = %d, want 6", summary.RentedRoomCount)
	}
	if summary.VacantRoomCount != 1 {
		t.Fatalf("vacant room count = %d, want 1", summary.VacantRoomCount)
	}
	if summary.RoomCount != 7 {
		t.Fatalf("room count = %d, want 7", summary.RoomCount)
	}
	assertHomeTodos(t, summary.Todos, []Todo{
		{Title: unit.Name, Description: "还有 1 间房未抄表", Tag: todoTagMeterReading},
		{Title: unit.Name, Description: "还有 2 间房待生成账单", Tag: todoTagBillGeneration},
		{Title: unit.Name, Description: "还有 1 笔租金待收", Tag: todoTagRentCollection},
	})
}

func TestGetHomeSummaryFiltersTodosByPermission(t *testing.T) {
	periodMonth := time.Now().Format("2006-01")
	unit := &domainunit.Unit{ID: 1, AccountID: 20, Name: "阳光公寓 · 1栋"}
	water := &domainmeteritem.MeterItem{ID: 1}
	rooms := []*domainroom.Room{
		{ID: 11, UnitID: 1, IsActive: false, MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 12, UnitID: 1, IsActive: false},
		{ID: 13, UnitID: 1, IsActive: false},
	}
	bills := []*domainrentbill.RentBill{
		{ID: 1, RoomID: 13, PeriodMonth: periodMonth, Status: domainrentbill.StatusUnpaid, TotalAmount: 1000},
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), 30)

	t.Run("hide meter reading todos", func(t *testing.T) {
		service := newHomeTestService(
			[]*domainunit.Unit{unit},
			rooms,
			nil,
			bills,
			&fakeHomeRbacService{denied: map[string]error{domainrbac.PermissionMeterReadingView: domainrbac.ErrForbidden}},
		)

		summary, err := service.GetHomeSummary(ctx)
		if err != nil {
			t.Fatalf("GetHomeSummary() error = %v", err)
		}
		assertHomeTodos(t, summary.Todos, []Todo{
			{Title: unit.Name, Description: "还有 1 间房待生成账单", Tag: todoTagBillGeneration},
			{Title: unit.Name, Description: "还有 1 笔租金待收", Tag: todoTagRentCollection},
		})
	})

	t.Run("hide rent bill todos", func(t *testing.T) {
		service := newHomeTestService(
			[]*domainunit.Unit{unit},
			rooms,
			nil,
			bills,
			&fakeHomeRbacService{denied: map[string]error{domainrbac.PermissionRentBillView: domainrbac.ErrForbidden}},
		)

		summary, err := service.GetHomeSummary(ctx)
		if err != nil {
			t.Fatalf("GetHomeSummary() error = %v", err)
		}
		assertHomeTodos(t, summary.Todos, []Todo{
			{Title: unit.Name, Description: "还有 1 间房未抄表", Tag: todoTagMeterReading},
		})
	})
}

func TestHomeTodoBucketsLimitAndOrder(t *testing.T) {
	buckets := &homeTodoBuckets{}
	for i := 0; i < 5; i++ {
		buckets.meterReadings = append(buckets.meterReadings, Todo{Tag: todoTagMeterReading})
		buckets.billGenerations = append(buckets.billGenerations, Todo{Tag: todoTagBillGeneration})
		buckets.rentCollections = append(buckets.rentCollections, Todo{Tag: todoTagRentCollection})
	}

	todos := buckets.todos(10)
	if len(todos) != 10 {
		t.Fatalf("todo count = %d, want 10", len(todos))
	}
	for i := 0; i < 5; i++ {
		if todos[i].Tag != todoTagMeterReading {
			t.Fatalf("todo[%d].tag = %q, want %q", i, todos[i].Tag, todoTagMeterReading)
		}
	}
	for i := 5; i < 10; i++ {
		if todos[i].Tag != todoTagBillGeneration {
			t.Fatalf("todo[%d].tag = %q, want %q", i, todos[i].Tag, todoTagBillGeneration)
		}
	}
}

func assertHomeTodos(t *testing.T, got []Todo, want []Todo) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("todo count = %d, want %d, got %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("todo[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func newHomeTestService(units []*domainunit.Unit, rooms []*domainroom.Room, readings []*domainmeterreading.MeterReading, bills []*domainrentbill.RentBill, rbac domainrbac.IRbacService) *Service {
	roomUnitIDs := make(map[int]int, len(rooms))
	for _, room := range rooms {
		if room == nil {
			continue
		}
		roomUnitIDs[room.ID] = room.UnitID
	}
	return &Service{
		unitRepo:         &fakeHomeUnitRepo{units: units},
		roomRepo:         &fakeHomeRoomRepo{rooms: rooms},
		rentBillRepo:     &fakeHomeRentBillRepo{bills: bills, roomUnitIDs: roomUnitIDs},
		meterReadingRepo: &fakeHomeMeterReadingRepo{readings: readings, roomUnitIDs: roomUnitIDs},
		rbacService:      rbac,
	}
}

type fakeHomeUnitRepo struct {
	units []*domainunit.Unit
}

func (r *fakeHomeUnitRepo) Create(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeHomeUnitRepo) Get(ctx context.Context, accountID int, id int) (*domainunit.Unit, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeHomeUnitRepo) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	for _, unit := range r.units {
		if unit != nil && unit.ID == id {
			return unit, nil
		}
	}
	return nil, domainunit.ErrNotFound
}

func (r *fakeHomeUnitRepo) List(ctx context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	return append([]*domainunit.Unit(nil), r.units...), nil
}

func (r *fakeHomeUnitRepo) Update(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeHomeUnitRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeHomeRoomRepo struct {
	rooms []*domainroom.Room
}

func (r *fakeHomeRoomRepo) Create(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	return room, nil
}

func (r *fakeHomeRoomRepo) Get(ctx context.Context, accountID int, id int) (*domainroom.Room, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeHomeRoomRepo) GetByID(ctx context.Context, id int) (*domainroom.Room, error) {
	for _, room := range r.rooms {
		if room != nil && room.ID == id {
			return room, nil
		}
	}
	return nil, domainroom.ErrNotFound
}

func (r *fakeHomeRoomRepo) List(ctx context.Context, accountID int, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	rooms := make([]*domainroom.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		if room == nil || room.UnitID != query.UnitID {
			continue
		}
		if query.IsActive != nil && room.IsActive != *query.IsActive {
			continue
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *fakeHomeRoomRepo) Count(ctx context.Context, accountID int, query domainroom.ListQuery) (int, error) {
	rooms, err := r.List(ctx, accountID, query)
	return len(rooms), err
}

func (r *fakeHomeRoomRepo) Update(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	return room, nil
}

func (r *fakeHomeRoomRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeHomeRentBillRepo struct {
	bills       []*domainrentbill.RentBill
	roomUnitIDs map[int]int
}

func (r *fakeHomeRentBillRepo) Create(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	return bill, nil
}

func (r *fakeHomeRentBillRepo) Get(ctx context.Context, accountID int, id int) (*domainrentbill.RentBill, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeHomeRentBillRepo) GetByID(ctx context.Context, id int) (*domainrentbill.RentBill, error) {
	for _, bill := range r.bills {
		if bill != nil && bill.ID == id {
			return bill, nil
		}
	}
	return nil, domainrentbill.ErrNotFound
}

func (r *fakeHomeRentBillRepo) GetByRoomPeriod(ctx context.Context, accountID int, roomID int, periodMonth string) (*domainrentbill.RentBill, error) {
	for _, bill := range r.bills {
		if bill != nil && bill.RoomID == roomID && bill.PeriodMonth == periodMonth {
			return bill, nil
		}
	}
	return nil, domainrentbill.ErrNotFound
}

func (r *fakeHomeRentBillRepo) List(ctx context.Context, accountID int, query domainrentbill.ListQuery) ([]*domainrentbill.RentBill, error) {
	bills := make([]*domainrentbill.RentBill, 0, len(r.bills))
	for _, bill := range r.bills {
		if bill == nil {
			continue
		}
		if query.UnitID > 0 && r.roomUnitIDs[bill.RoomID] != query.UnitID {
			continue
		}
		if query.RoomID > 0 && bill.RoomID != query.RoomID {
			continue
		}
		if query.PeriodMonth != "" && bill.PeriodMonth != query.PeriodMonth {
			continue
		}
		if query.Status != "" && bill.Status != query.Status {
			continue
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

func (r *fakeHomeRentBillRepo) SaveGeneratedBills(ctx context.Context, accountID int, bills []*domainrentbill.RentBill) (*domainrentbill.SaveGeneratedBillsResult, error) {
	return &domainrentbill.SaveGeneratedBillsResult{Saved: bills}, nil
}

func (r *fakeHomeRentBillRepo) Update(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	return bill, nil
}

func (r *fakeHomeRentBillRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeHomeMeterReadingRepo struct {
	readings    []*domainmeterreading.MeterReading
	roomUnitIDs map[int]int
}

func (r *fakeHomeMeterReadingRepo) Create(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	return reading, nil
}

func (r *fakeHomeMeterReadingRepo) Get(ctx context.Context, accountID int, id int) (*domainmeterreading.MeterReading, error) {
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeHomeMeterReadingRepo) GetByRoomMeterItemPeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	for _, reading := range r.readings {
		if reading != nil && reading.RoomID == roomID && reading.MeterItemID == meterItemID && reading.PeriodMonth == periodMonth {
			return reading, nil
		}
	}
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeHomeMeterReadingRepo) GetLatestBeforePeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	return nil, domainmeterreading.ErrNotFound
}

func (r *fakeHomeMeterReadingRepo) List(ctx context.Context, accountID int, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	readings := make([]*domainmeterreading.MeterReading, 0, len(r.readings))
	for _, reading := range r.readings {
		if reading == nil {
			continue
		}
		if query.UnitID > 0 && r.roomUnitIDs[reading.RoomID] != query.UnitID {
			continue
		}
		if query.RoomID > 0 && reading.RoomID != query.RoomID {
			continue
		}
		if query.MeterItemID > 0 && reading.MeterItemID != query.MeterItemID {
			continue
		}
		if query.PeriodMonth != "" && reading.PeriodMonth != query.PeriodMonth {
			continue
		}
		readings = append(readings, reading)
	}
	return readings, nil
}

func (r *fakeHomeMeterReadingRepo) SaveRoomMonth(ctx context.Context, accountID int, readings []*domainmeterreading.MeterReading) ([]*domainmeterreading.MeterReading, error) {
	return readings, nil
}

func (r *fakeHomeMeterReadingRepo) Update(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	return reading, nil
}

func (r *fakeHomeMeterReadingRepo) Delete(ctx context.Context, accountID int, id int) error {
	return nil
}

type fakeHomeRbacService struct {
	denied map[string]error
}

func (s *fakeHomeRbacService) RequireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) error {
	if s.denied == nil {
		return nil
	}
	if err := s.denied[permissionCode]; err != nil {
		return err
	}
	return nil
}

var _ domainunit.Repository = (*fakeHomeUnitRepo)(nil)
var _ domainroom.Repository = (*fakeHomeRoomRepo)(nil)
var _ domainrentbill.Repository = (*fakeHomeRentBillRepo)(nil)
var _ domainmeterreading.Repository = (*fakeHomeMeterReadingRepo)(nil)
var _ domainrbac.IRbacService = (*fakeHomeRbacService)(nil)
