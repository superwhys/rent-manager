package meterreading

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestGetUnitMonthReadingsDefaultHidesVacantRecords(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200, CalcType: domainmeteritem.CalcTypeDirect}
	rooms := []*domainroom.Room{
		{ID: 11, UnitID: 7, RoomNo: "101", MeterItems: []*domainmeteritem.MeterItem{water}},
		{ID: 12, UnitID: 7, RoomNo: "301", IsActive: true, MeterItems: []*domainmeteritem.MeterItem{water}},
	}
	service, roomRepo, readingRepo := newRoomScopeService(rooms, []*domainmeterreading.MeterReading{
		roomScopeReading(11, 1, "2026-10", 3),
		roomScopeReading(12, 1, "2026-10", 4),
	})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
	unitMonth, err := service.GetUnitMonthReadings(ctx, &UnitMeterReadingsQuery{UnitID: 7, PeriodMonth: "2026-10"})
	if err != nil {
		t.Fatalf("GetUnitMonthReadings() error = %v", err)
	}
	if len(unitMonth.Rooms) != 1 || unitMonth.Rooms[0].Room.ID != 11 {
		t.Fatalf("displayed rooms = %v, want only rented room 101", unitMonth.Rooms)
	}
	if unitMonth.MissingCount != 0 || unitMonth.TotalFee != 600 || len(unitMonth.Summaries) != 1 {
		t.Fatalf("unit month = %+v, want no missing, fee 600 and one summary", unitMonth)
	}
	summary := unitMonth.Summaries[0]
	if summary.UsageAmount != 3 || summary.Fee != 600 || summary.Recorded != 1 || summary.Missing != 0 {
		t.Fatalf("summary = %+v, want only rented room usage 3 and fee 600", summary)
	}
	if roomRepo.query.IsActive == nil || *roomRepo.query.IsActive {
		t.Fatalf("room query = %+v, want rented rooms only", roomRepo.query)
	}
	for _, query := range readingRepo.queries {
		if query.RoomID == 12 {
			t.Fatalf("vacant room readings were queried: %+v", query)
		}
	}
}

func TestGetUnitMonthReadingsRoomScope(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200, CalcType: domainmeteritem.CalcTypeDirect}
	power := &domainmeteritem.MeterItem{ID: 2, Code: "电", Unit: "kWh", UnitPrice: 300, CalcType: domainmeteritem.CalcTypeDirect}
	tests := []struct {
		name        string
		isActive    bool
		readings    []*domainmeterreading.MeterReading
		wantMissing int
		wantFee     int
	}{
		{name: "已出租无记录仍显示并待录入", wantMissing: 2},
		{name: "空房无记录显示且无待录入", isActive: true},
		{name: "空房仅其他月份有记录仍显示且无待录入", isActive: true, readings: []*domainmeterreading.MeterReading{roomScopeReading(11, 1, "2026-09", 3)}},
		{name: "空房零用量记录保留且无待录入", isActive: true, readings: []*domainmeterreading.MeterReading{roomScopeReading(11, 1, "2026-10", 0)}},
		{name: "空房部分记录保留且无待录入", isActive: true, readings: []*domainmeterreading.MeterReading{roomScopeReading(11, 1, "2026-10", 3)}, wantFee: 600},
		{name: "空房完整记录保留费用", isActive: true, readings: []*domainmeterreading.MeterReading{roomScopeReading(11, 1, "2026-10", 3), roomScopeReading(11, 2, "2026-10", 4)}, wantFee: 1800},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := &domainroom.Room{ID: 11, UnitID: 7, RoomNo: "301", IsActive: tt.isActive, MeterItems: []*domainmeteritem.MeterItem{water, power}}
			service, _, _ := newRoomScopeService([]*domainroom.Room{room}, tt.readings)
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
			unitMonth, err := service.GetUnitMonthReadings(ctx, &UnitMeterReadingsQuery{UnitID: 7, PeriodMonth: "2026-10", IncludeVacantRecords: true})
			if err != nil {
				t.Fatalf("GetUnitMonthReadings() error = %v", err)
			}
			if len(unitMonth.Rooms) != 1 {
				t.Fatalf("room count = %d, want 1", len(unitMonth.Rooms))
			}
			if unitMonth.MissingCount != tt.wantMissing || unitMonth.TotalFee != tt.wantFee {
				t.Fatalf("unit missing = %d, fee = %d, want %d, %d", unitMonth.MissingCount, unitMonth.TotalFee, tt.wantMissing, tt.wantFee)
			}
			roomMonth := unitMonth.Rooms[0]
			if roomMonth.MissingCount != tt.wantMissing || roomMonth.TotalFee != tt.wantFee {
				t.Fatalf("room missing = %d, fee = %d, want %d, %d", roomMonth.MissingCount, roomMonth.TotalFee, tt.wantMissing, tt.wantFee)
			}
			encoded, err := json.Marshal(ToRoomMonthReadingsResponse(roomMonth))
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(encoded, &payload); err != nil {
				t.Fatal(err)
			}
			if payload["is_active"] != tt.isActive {
				t.Fatalf("response is_active = %v, want %v", payload["is_active"], tt.isActive)
			}
			if tt.isActive {
				for _, summary := range unitMonth.Summaries {
					if summary.Missing != 0 {
						t.Fatalf("vacant room item %d missing = %d, want 0", summary.MeterItem.ID, summary.Missing)
					}
				}
				for _, reading := range tt.readings {
					if reading.PeriodMonth != unitMonth.PeriodMonth {
						continue
					}
					found := false
					for _, line := range roomMonth.Readings {
						if line.MeterItem.ID == reading.MeterItemID && line.IsRecorded && line.Reading == reading {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("recorded item %d is absent from retained room", reading.MeterItemID)
					}
				}
			}
		})
	}
}

func TestGetUnitMonthReadingsRoomScopeSummary(t *testing.T) {
	water := &domainmeteritem.MeterItem{ID: 1, Code: "水", Unit: "m3", UnitPrice: 200, CalcType: domainmeteritem.CalcTypeDirect}
	power := &domainmeteritem.MeterItem{ID: 2, Code: "电", Unit: "kWh", UnitPrice: 300, CalcType: domainmeteritem.CalcTypeDirect}
	rooms := []*domainroom.Room{
		{ID: 11, UnitID: 7, RoomNo: "101", MeterItems: []*domainmeteritem.MeterItem{water, power}},
		{ID: 12, UnitID: 7, RoomNo: "102", MeterItems: []*domainmeteritem.MeterItem{water, power}},
		{ID: 13, UnitID: 7, RoomNo: "301", IsActive: true, MeterItems: []*domainmeteritem.MeterItem{water, power}},
		{ID: 14, UnitID: 7, RoomNo: "302", IsActive: true, MeterItems: []*domainmeteritem.MeterItem{water, power}},
		{ID: 15, UnitID: 7, RoomNo: "303", IsActive: true, MeterItems: []*domainmeteritem.MeterItem{water, power}},
		{ID: 16, UnitID: 8, RoomNo: "其他楼栋", MeterItems: []*domainmeteritem.MeterItem{water, power}},
	}
	service, roomRepo, readingRepo := newRoomScopeService(rooms, []*domainmeterreading.MeterReading{
		roomScopeReading(12, 1, "2026-10", 3),
		roomScopeReading(14, 2, "2026-10", 4),
		roomScopeReading(15, 1, "2026-09", 5),
		roomScopeReading(16, 1, "2026-10", 99),
	})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	unitMonth, err := service.GetUnitMonthReadings(ctx, &UnitMeterReadingsQuery{UnitID: 7, PeriodMonth: "2026-10", IncludeVacantRecords: true})
	if err != nil {
		t.Fatalf("GetUnitMonthReadings() error = %v", err)
	}
	if len(unitMonth.Rooms) != 5 || unitMonth.Rooms[0].Room.ID != 11 || unitMonth.Rooms[1].Room.ID != 12 || unitMonth.Rooms[2].Room.ID != 13 || unitMonth.Rooms[3].Room.ID != 14 || unitMonth.Rooms[4].Room.ID != 15 {
		t.Fatalf("displayed rooms = %v, want 101, 102, 301, 302, 303", unitMonth.Rooms)
	}
	if unitMonth.MissingCount != 3 || unitMonth.TotalFee != 1800 {
		t.Fatalf("unit missing = %d, fee = %d, want 3, 1800", unitMonth.MissingCount, unitMonth.TotalFee)
	}
	if len(unitMonth.Summaries) != 2 {
		t.Fatalf("summary count = %d, want 2", len(unitMonth.Summaries))
	}
	for _, summary := range unitMonth.Summaries {
		wantMissing, wantUsage, wantFee := 1, 3.0, 600
		if summary.MeterItem.ID == 2 {
			wantMissing, wantUsage, wantFee = 2, 4, 1200
		}
		if summary.Missing != wantMissing || summary.Recorded != 1 || summary.UsageAmount != wantUsage || summary.Fee != wantFee {
			t.Fatalf("item %d summary = %+v, want missing %d, recorded 1, usage %v, fee %d", summary.MeterItem.ID, summary, wantMissing, wantUsage, wantFee)
		}
	}
	if roomRepo.accountID != testAccountID20 || roomRepo.query.UnitID != 7 || roomRepo.query.IsActive != nil {
		t.Fatalf("room query account = %q, query = %+v, want owner's unit and all occupancy states", roomRepo.accountID, roomRepo.query)
	}
	if len(readingRepo.queries) == 0 {
		t.Fatal("no reading query was issued")
	}
	for _, query := range readingRepo.queries {
		if query.PeriodMonth != "2026-10" || (query.RoomID == 0 && query.UnitID != 7) {
			t.Fatalf("reading query = %+v, want selected month scoped to unit or its rooms", query)
		}
	}
	rbac := service.rbacService.(*fakeRbacService)
	if rbac.accountID != testAccountID10 || rbac.unitID != 7 || rbac.permissionCode != domainrbac.PermissionMeterReadingView {
		t.Fatalf("RBAC check = %+v, want caller's meter_reading.view permission for unit 7", rbac)
	}
}

func TestGetUnitMonthReadingsVacantRoomWithUnassignedReadingIsRetained(t *testing.T) {
	room := &domainroom.Room{ID: 11, UnitID: 7, RoomNo: "301", IsActive: true}
	service, _, _ := newRoomScopeService([]*domainroom.Room{room}, []*domainmeterreading.MeterReading{
		roomScopeReading(11, 1, "2026-10", 3),
	})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
	unitMonth, err := service.GetUnitMonthReadings(ctx, &UnitMeterReadingsQuery{UnitID: 7, PeriodMonth: "2026-10", IncludeVacantRecords: true})
	if err != nil {
		t.Fatalf("GetUnitMonthReadings() error = %v", err)
	}
	if len(unitMonth.Rooms) != 1 || unitMonth.Rooms[0].Room.ID != 11 {
		t.Fatalf("displayed rooms = %v, want room 301 with a saved reading", unitMonth.Rooms)
	}
	if unitMonth.MissingCount != 0 || unitMonth.Rooms[0].MissingCount != 0 {
		t.Fatalf("vacant room missing = %d, unit missing = %d, want 0", unitMonth.Rooms[0].MissingCount, unitMonth.MissingCount)
	}
}

func TestGetUnitMonthReadingsRoomScopeRequiresPermission(t *testing.T) {
	service, roomRepo, readingRepo := newRoomScopeService(nil, nil)
	service.rbacService = &fakeRbacService{err: domainrbac.ErrForbidden}
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	_, err := service.GetUnitMonthReadings(ctx, &UnitMeterReadingsQuery{UnitID: 7, PeriodMonth: "2026-10"})
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
	if roomRepo.accountID != "" || len(readingRepo.queries) != 0 {
		t.Fatal("room or reading data was queried before permission was granted")
	}
}

func roomScopeReading(roomID, itemID int, periodMonth string, usage float64) *domainmeterreading.MeterReading {
	return &domainmeterreading.MeterReading{ID: roomID*10 + itemID, RoomID: roomID, MeterItemID: itemID, PeriodMonth: periodMonth, CurrentReading: usage, UsageAmount: &usage}
}

func newRoomScopeService(rooms []*domainroom.Room, readings []*domainmeterreading.MeterReading) (*Service, *roomScopeRoomRepo, *roomScopeReadingRepo) {
	roomRepo := &roomScopeRoomRepo{rooms: rooms}
	readingRepo := &roomScopeReadingRepo{rooms: rooms, readings: readings}
	service := &Service{
		repo: readingRepo, roomRepo: roomRepo,
		unitRepo:    &fakeUnitRepo{byID: map[int]*domainunit.Unit{7: {ID: 7, AccountID: testAccountID20}}},
		rbacService: &fakeRbacService{},
	}
	return service, roomRepo, readingRepo
}

type roomScopeRoomRepo struct {
	fakeRoomRepo
	rooms     []*domainroom.Room
	accountID string
	query     domainroom.ListQuery
}

func (r *roomScopeRoomRepo) List(ctx context.Context, accountID string, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	r.accountID, r.query = accountID, query
	if accountID != testAccountID20 {
		return nil, domainrbac.ErrForbidden
	}
	rooms := make([]*domainroom.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		if room.UnitID == query.UnitID && (query.IsActive == nil || room.IsActive == *query.IsActive) {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

type roomScopeReadingRepo struct {
	fakeMeterReadingRepo
	rooms    []*domainroom.Room
	readings []*domainmeterreading.MeterReading
	queries  []domainmeterreading.ListQuery
}

func (r *roomScopeReadingRepo) List(ctx context.Context, accountID string, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	if accountID != testAccountID20 {
		return nil, domainrbac.ErrForbidden
	}
	r.queries = append(r.queries, query)
	readings := make([]*domainmeterreading.MeterReading, 0)
	for _, reading := range r.readings {
		if query.RoomID != 0 && reading.RoomID != query.RoomID {
			continue
		}
		if query.UnitID != 0 {
			inUnit := false
			for _, room := range r.rooms {
				if room.ID == reading.RoomID && room.UnitID == query.UnitID {
					inUnit = true
					break
				}
			}
			if !inUnit {
				continue
			}
		}
		if reading.PeriodMonth == query.PeriodMonth {
			readings = append(readings, reading)
		}
	}
	return readings, nil
}
