package services

import (
	"context"
	"errors"
	"testing"

	domainrbac "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/rbac"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/models/requests"
)

func TestRoomServiceListRoomsAllowsOwnerWithoutRBAC(t *testing.T) {
	roomRepo := &fakeRoomRepository{}
	service := NewRoomService(roomRepo, &fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}, nil)

	_, err := service.ListRoomsInUnit(tokenContext(1), &requests.UnitRoomsQuery{UnitID: 10})
	if err != nil {
		t.Fatalf("owner should list own unit rooms without rbac, got %v", err)
	}
	if roomRepo.listAccountID != 1 || roomRepo.listQuery.UnitID != 10 {
		t.Fatalf("unexpected list query: account=%d query=%#v", roomRepo.listAccountID, roomRepo.listQuery)
	}
}

func TestRoomServiceListRoomsRequiresViewPermissionForGrantedUnit(t *testing.T) {
	roomRepo := &fakeRoomRepository{}
	unitRepo := &fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}
	service := NewRoomService(roomRepo, unitRepo, NewRBACService(&fakeRBACRepository{}, nil))

	_, err := service.ListRoomsInUnit(tokenContext(2), &requests.UnitRoomsQuery{UnitID: 10})
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without room.view, got %v", err)
	}

	service = NewRoomService(roomRepo, unitRepo, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionRoomView: true}},
		},
	}, nil))
	_, err = service.ListRoomsInUnit(tokenContext(2), &requests.UnitRoomsQuery{UnitID: 10})
	if err != nil {
		t.Fatalf("expected granted room.view, got %v", err)
	}
}

func TestRoomServiceUpdateRoomRequiresManagePermissionForGrantedUnit(t *testing.T) {
	roomRepo := &fakeRoomRepository{
		rooms: map[int]*domainroom.Room{
			20: {ID: 20, RoomNo: "101", Floor: 1, Price: 1000, Size: 20, UnitID: 10, RoomTypeID: 1, IsActive: true},
		},
	}
	unitRepo := &fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}

	service := NewRoomService(roomRepo, unitRepo, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionRoomView: true}},
		},
	}, nil))
	_, err := service.UpdateRoom(tokenContext(2), &requests.UpdateRoomCommand{RoomID: 20, RoomNo: "102"})
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without room.manage, got %v", err)
	}

	service = NewRoomService(roomRepo, unitRepo, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionRoomManage: true}},
		},
	}, nil))
	room, err := service.UpdateRoom(tokenContext(2), &requests.UpdateRoomCommand{RoomID: 20, RoomNo: "102"})
	if err != nil {
		t.Fatalf("expected granted room.manage, got %v", err)
	}
	if room.RoomNo != "102" {
		t.Fatalf("expected updated room no, got %q", room.RoomNo)
	}
}

type fakeUnitRepository struct {
	units     map[int]*domainunit.Unit
	listUnits []*domainunit.Unit
	listQuery domainunit.ListQuery
}

func (f *fakeUnitRepository) Create(context.Context, *domainunit.Unit) (*domainunit.Unit, error) {
	return nil, nil
}

func (f *fakeUnitRepository) Get(_ context.Context, accountID int, id int) (*domainunit.Unit, error) {
	unit, err := f.GetByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if unit.AccountID != accountID {
		return nil, domainunit.ErrNotFound
	}
	return unit, nil
}

func (f *fakeUnitRepository) GetByID(_ context.Context, id int) (*domainunit.Unit, error) {
	unit, ok := f.units[id]
	if !ok {
		return nil, domainunit.ErrNotFound
	}
	return unit, nil
}

func (f *fakeUnitRepository) List(_ context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	f.listQuery = query
	return f.listUnits, nil
}

func (f *fakeUnitRepository) Update(_ context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	if f.units == nil {
		f.units = make(map[int]*domainunit.Unit)
	}
	copied := *unit
	f.units[unit.ID] = &copied
	return &copied, nil
}

func (f *fakeUnitRepository) Delete(context.Context, int, int) error {
	return nil
}

type fakeRoomRepository struct {
	rooms         map[int]*domainroom.Room
	listAccountID int
	listQuery     domainroom.ListQuery
	countCalls    []roomCountCall
}

type roomCountCall struct {
	accountID int
	query     domainroom.ListQuery
}

func (f *fakeRoomRepository) Create(context.Context, int, *domainroom.Room) (*domainroom.Room, error) {
	return nil, nil
}

func (f *fakeRoomRepository) Get(context.Context, int, int) (*domainroom.Room, error) {
	return nil, nil
}

func (f *fakeRoomRepository) GetByID(_ context.Context, id int) (*domainroom.Room, error) {
	room, ok := f.rooms[id]
	if !ok {
		return nil, domainroom.ErrNotFound
	}
	copied := *room
	return &copied, nil
}

func (f *fakeRoomRepository) List(_ context.Context, accountID int, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	f.listAccountID = accountID
	f.listQuery = query
	return nil, nil
}

func (f *fakeRoomRepository) Count(_ context.Context, accountID int, query domainroom.ListQuery) (int, error) {
	f.countCalls = append(f.countCalls, roomCountCall{accountID: accountID, query: query})
	count := 0
	for _, room := range f.rooms {
		if room.UnitID != query.UnitID {
			continue
		}
		if query.RoomTypeID > 0 && room.RoomTypeID != query.RoomTypeID {
			continue
		}
		if query.IsActive != nil && room.IsActive != *query.IsActive {
			continue
		}
		count++
	}
	return count, nil
}

func (f *fakeRoomRepository) Update(_ context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	f.rooms[room.ID] = room
	return room, nil
}

func (f *fakeRoomRepository) Delete(context.Context, int, int) error {
	return nil
}
