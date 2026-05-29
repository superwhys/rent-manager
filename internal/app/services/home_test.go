package services

import (
	"testing"

	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
)

func TestHomeServiceGetHomeSummaryAggregatesAccessibleUnits(t *testing.T) {
	unitRepo := &fakeUnitRepository{
		listUnits: []*domainunit.Unit{
			{ID: 10, AccountID: 1, Name: "自有楼栋"},
			{ID: 20, AccountID: 3, Name: "授权楼栋"},
		},
	}
	roomRepo := &fakeRoomRepository{
		rooms: map[int]*domainroom.Room{
			1: {ID: 1, UnitID: 10, RoomNo: "101", IsActive: true},
			2: {ID: 2, UnitID: 10, RoomNo: "102", IsActive: false},
			3: {ID: 3, UnitID: 20, RoomNo: "201", IsActive: true},
			4: {ID: 4, UnitID: 20, RoomNo: "202", IsActive: false},
		},
	}
	service := NewHomeService(unitRepo, roomRepo)

	got, err := service.GetHomeSummary(tokenContext(1))
	if err != nil {
		t.Fatalf("GetHomeSummary() error = %v", err)
	}

	if got.BuildingCount != 2 {
		t.Fatalf("expected 2 buildings, got %d", got.BuildingCount)
	}
	if got.RoomCount != 4 || got.RentedRoomCount != 2 || got.VacantRoomCount != 2 {
		t.Fatalf("unexpected room summary: %#v", got)
	}
	if unitRepo.listQuery.AccountID != 1 || unitRepo.listQuery.Scope != domainunit.ListScopeAll {
		t.Fatalf("unexpected unit list query: %#v", unitRepo.listQuery)
	}
	if len(roomRepo.countCalls) != 4 {
		t.Fatalf("expected 4 room count calls, got %d", len(roomRepo.countCalls))
	}
	if roomRepo.countCalls[0].accountID != 1 || roomRepo.countCalls[0].query.UnitID != 10 {
		t.Fatalf("unexpected first room count call: %#v", roomRepo.countCalls[0])
	}
	if roomRepo.countCalls[2].accountID != 3 || roomRepo.countCalls[2].query.UnitID != 20 {
		t.Fatalf("unexpected granted unit room count call: %#v", roomRepo.countCalls[2])
	}
}
