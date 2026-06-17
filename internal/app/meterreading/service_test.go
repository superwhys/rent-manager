package meterreading

import (
	"context"
	"errors"
	"testing"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
	domainmeterreading "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meterreading"
	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
)

func TestSaveRoomMonthReadingsUsesLatestPreviousReading(t *testing.T) {
	item := &domainmeteritem.MeterItem{
		ID:        1,
		Code:      "水",
		Unit:      "m3",
		UnitPrice: 200,
		CalcType:  domainmeteritem.CalcTypeDifference,
	}
	room := &domainroom.Room{ID: 11, UnitID: 7, RoomNo: "101", MeterItems: []*domainmeteritem.MeterItem{item}}
	unit := &domainunit.Unit{ID: 7, AccountID: 10}
	repo := &fakeMeterReadingRepo{
		latestByMeterItemID: map[int]*domainmeterreading.MeterReading{
			1: {RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-05", CurrentReading: 100},
		},
	}
	rbac := &fakeRbacService{}
	service := &Service{
		repo:        repo,
		roomRepo:    &fakeRoomRepo{byID: map[int]*domainroom.Room{11: room}},
		unitRepo:    &fakeUnitRepo{byID: map[int]*domainunit.Unit{7: unit}},
		rbacService: rbac,
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)
	inputPrevious := 1.0

	roomMonth, err := service.SaveRoomMonthReadings(ctx, &SaveRoomMonthCommand{
		RoomID:      11,
		PeriodMonth: "2026-06",
		Readings: []MeterReadingInput{
			{MeterItemID: 1, PreviousReading: &inputPrevious, CurrentReading: 123},
		},
	})
	if err != nil {
		t.Fatalf("SaveRoomMonthReadings() error = %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved count = %d, want 1", len(repo.saved))
	}
	saved := repo.saved[0]
	if saved.PreviousReading == nil || *saved.PreviousReading != 100 {
		t.Fatalf("previous reading = %v, want 100", saved.PreviousReading)
	}
	if saved.UsageAmount == nil || *saved.UsageAmount != 23 {
		t.Fatalf("usage amount = %v, want 23", saved.UsageAmount)
	}
	if saved.ReaderID == nil || *saved.ReaderID != 20 {
		t.Fatalf("reader id = %v, want 20", saved.ReaderID)
	}
	if roomMonth.TotalFee != 4600 {
		t.Fatalf("total fee = %d, want 4600", roomMonth.TotalFee)
	}
	if rbac.permissionCode != domainrbac.PermissionMeterReadingManage {
		t.Fatalf("permission = %q, want %q", rbac.permissionCode, domainrbac.PermissionMeterReadingManage)
	}
}

func TestSaveRoomMonthReadingsRequiresFirstBaseline(t *testing.T) {
	item := &domainmeteritem.MeterItem{
		ID:        1,
		Code:      "水",
		Unit:      "m3",
		UnitPrice: 200,
		CalcType:  domainmeteritem.CalcTypeDifference,
	}
	room := &domainroom.Room{ID: 11, UnitID: 7, RoomNo: "101", MeterItems: []*domainmeteritem.MeterItem{item}}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	repo := &fakeMeterReadingRepo{}
	service := &Service{
		repo:        repo,
		roomRepo:    &fakeRoomRepo{byID: map[int]*domainroom.Room{11: room}},
		unitRepo:    &fakeUnitRepo{byID: map[int]*domainunit.Unit{7: unit}},
		rbacService: &fakeRbacService{},
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)

	_, err := service.SaveRoomMonthReadings(ctx, &SaveRoomMonthCommand{
		RoomID:      11,
		PeriodMonth: "2026-06",
		Readings: []MeterReadingInput{
			{MeterItemID: 1, CurrentReading: 123},
		},
	})
	if !errors.Is(err, domainmeterreading.ErrPreviousReadingRequired) {
		t.Fatalf("error = %v, want ErrPreviousReadingRequired", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved count = %d, want 0", len(repo.saved))
	}
}

func TestSaveRoomMonthReadingsRejectsUnassignedMeterItem(t *testing.T) {
	item := &domainmeteritem.MeterItem{
		ID:        1,
		Code:      "水",
		Unit:      "m3",
		UnitPrice: 200,
		CalcType:  domainmeteritem.CalcTypeDifference,
	}
	room := &domainroom.Room{ID: 11, UnitID: 7, RoomNo: "101", MeterItems: []*domainmeteritem.MeterItem{item}}
	unit := &domainunit.Unit{ID: 7, AccountID: 20}
	repo := &fakeMeterReadingRepo{}
	service := &Service{
		repo:        repo,
		roomRepo:    &fakeRoomRepo{byID: map[int]*domainroom.Room{11: room}},
		unitRepo:    &fakeUnitRepo{byID: map[int]*domainunit.Unit{7: unit}},
		rbacService: &fakeRbacService{},
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), 20)
	previous := 100.0

	_, err := service.SaveRoomMonthReadings(ctx, &SaveRoomMonthCommand{
		RoomID:      11,
		PeriodMonth: "2026-06",
		Readings: []MeterReadingInput{
			{MeterItemID: 2, PreviousReading: &previous, CurrentReading: 123},
		},
	})
	if !errors.Is(err, domainmeterreading.ErrRoomMeterItemNotAssigned) {
		t.Fatalf("error = %v, want ErrRoomMeterItemNotAssigned", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved count = %d, want 0", len(repo.saved))
	}
}

type fakeMeterReadingRepo struct {
	listByRoomID        map[int][]*domainmeterreading.MeterReading
	latestByMeterItemID map[int]*domainmeterreading.MeterReading
	saved               []*domainmeterreading.MeterReading
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
	reading := r.latestByMeterItemID[meterItemID]
	if reading == nil {
		return nil, domainmeterreading.ErrNotFound
	}
	return reading, nil
}

func (r *fakeMeterReadingRepo) List(ctx context.Context, accountID int, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	if r.listByRoomID == nil {
		return nil, nil
	}
	return r.listByRoomID[query.RoomID], nil
}

func (r *fakeMeterReadingRepo) SaveRoomMonth(ctx context.Context, accountID int, readings []*domainmeterreading.MeterReading) ([]*domainmeterreading.MeterReading, error) {
	r.saved = append([]*domainmeterreading.MeterReading(nil), readings...)
	if r.listByRoomID == nil {
		r.listByRoomID = make(map[int][]*domainmeterreading.MeterReading)
	}
	for _, reading := range readings {
		if reading.ID == 0 {
			reading.ID = reading.MeterItemID
		}
		roomReadings := r.listByRoomID[reading.RoomID]
		replaced := false
		for index, existing := range roomReadings {
			if existing.MeterItemID == reading.MeterItemID && existing.PeriodMonth == reading.PeriodMonth {
				roomReadings[index] = reading
				replaced = true
				break
			}
		}
		if !replaced {
			roomReadings = append(roomReadings, reading)
		}
		r.listByRoomID[reading.RoomID] = roomReadings
	}
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
	list []*domainroom.Room
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
	if r.list != nil {
		return r.list, nil
	}
	rooms := make([]*domainroom.Room, 0, len(r.byID))
	for _, room := range r.byID {
		if room.UnitID == query.UnitID {
			rooms = append(rooms, room)
		}
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
	units := make([]*domainunit.Unit, 0, len(r.byID))
	for _, unit := range r.byID {
		units = append(units, unit)
	}
	return units, nil
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

var _ domainmeterreading.Repository = (*fakeMeterReadingRepo)(nil)
var _ domainroom.Repository = (*fakeRoomRepo)(nil)
var _ domainunit.Repository = (*fakeUnitRepo)(nil)
var _ domainrbac.IRbacService = (*fakeRbacService)(nil)
