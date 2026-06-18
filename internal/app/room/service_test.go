package room

import (
	"context"
	"slices"
	"testing"

	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
)

func TestCreateRoomInUnitBindsUnitDefaultMeterItems(t *testing.T) {
	service := newRoomTestService(
		[]*domainunit.Unit{
			{ID: 7, AccountID: 10, Name: "1栋", DefaultMeterItemIDs: []int{1, 2}},
		},
		nil,
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	room, err := service.CreateRoomInUnit(ctx, &CreateRoomInUnitCommand{
		UnitID:     7,
		RoomNo:     "101",
		Floor:      1,
		Price:      1000,
		Size:       20,
		RoomTypeID: 3,
	})
	if err != nil {
		t.Fatalf("CreateRoomInUnit() error = %v", err)
	}
	if !slices.Equal(room.MeterItemIDs, []int{1, 2}) {
		t.Fatalf("meter item ids = %#v, want [1 2]", room.MeterItemIDs)
	}
}

func TestCreateRoomInUnitMergesDefaultAndRequestMeterItems(t *testing.T) {
	service := newRoomTestService(
		[]*domainunit.Unit{
			{ID: 7, AccountID: 10, Name: "1栋", DefaultMeterItemIDs: []int{1, 2}},
		},
		nil,
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	room, err := service.CreateRoomInUnit(ctx, &CreateRoomInUnitCommand{
		UnitID:       7,
		RoomNo:       "101",
		Floor:        1,
		Price:        1000,
		Size:         20,
		RoomTypeID:   3,
		MeterItemIDs: []int{2, 3},
	})
	if err != nil {
		t.Fatalf("CreateRoomInUnit() error = %v", err)
	}
	if !slices.Equal(room.MeterItemIDs, []int{1, 2, 3}) {
		t.Fatalf("meter item ids = %#v, want [1 2 3]", room.MeterItemIDs)
	}
}

func TestUpdateRoomCanClearDefaultMeterItems(t *testing.T) {
	service := newRoomTestService(
		[]*domainunit.Unit{
			{ID: 7, AccountID: 10, Name: "1栋", DefaultMeterItemIDs: []int{1, 2}},
		},
		[]*domainroom.Room{
			{
				ID:           11,
				RoomNo:       "101",
				Floor:        1,
				Price:        1000,
				Size:         20,
				UnitID:       7,
				RoomTypeID:   3,
				IsActive:     false,
				MeterItemIDs: []int{1, 2},
			},
		},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), 10)

	room, err := service.UpdateRoom(ctx, &UpdateRoomCommand{
		RoomID:       11,
		MeterItemIDs: []int{},
	})
	if err != nil {
		t.Fatalf("UpdateRoom() error = %v", err)
	}
	if len(room.MeterItemIDs) != 0 {
		t.Fatalf("meter item ids = %#v, want empty", room.MeterItemIDs)
	}
}

func newRoomTestService(units []*domainunit.Unit, rooms []*domainroom.Room) *Service {
	unitByID := make(map[int]*domainunit.Unit, len(units))
	for _, unit := range units {
		if unit == nil {
			continue
		}
		unitByID[unit.ID] = unit
	}
	roomByID := make(map[int]*domainroom.Room, len(rooms))
	for _, room := range rooms {
		if room == nil {
			continue
		}
		roomByID[room.ID] = room
	}
	return &Service{
		repo:     &fakeRoomServiceRoomRepo{byID: roomByID, nextID: 100},
		unitRepo: &fakeRoomServiceUnitRepo{byID: unitByID},
	}
}

type fakeRoomServiceRoomRepo struct {
	byID   map[int]*domainroom.Room
	nextID int
}

func (r *fakeRoomServiceRoomRepo) Create(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if r.byID == nil {
		r.byID = make(map[int]*domainroom.Room)
	}
	created := cloneRoom(room)
	if created.ID == 0 {
		r.nextID++
		created.ID = r.nextID
	}
	r.byID[created.ID] = created
	return cloneRoom(created), nil
}

func (r *fakeRoomServiceRoomRepo) Get(ctx context.Context, accountID int, id int) (*domainroom.Room, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeRoomServiceRoomRepo) GetByID(ctx context.Context, id int) (*domainroom.Room, error) {
	room := r.byID[id]
	if room == nil {
		return nil, domainroom.ErrNotFound
	}
	return cloneRoom(room), nil
}

func (r *fakeRoomServiceRoomRepo) List(ctx context.Context, accountID int, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	rooms := make([]*domainroom.Room, 0, len(r.byID))
	for _, room := range r.byID {
		if room == nil || room.UnitID != query.UnitID {
			continue
		}
		rooms = append(rooms, cloneRoom(room))
	}
	return rooms, nil
}

func (r *fakeRoomServiceRoomRepo) Count(ctx context.Context, accountID int, query domainroom.ListQuery) (int, error) {
	rooms, err := r.List(ctx, accountID, query)
	return len(rooms), err
}

func (r *fakeRoomServiceRoomRepo) Update(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if r.byID == nil || r.byID[room.ID] == nil {
		return nil, domainroom.ErrNotFound
	}
	updated := cloneRoom(room)
	r.byID[updated.ID] = updated
	return cloneRoom(updated), nil
}

func (r *fakeRoomServiceRoomRepo) Delete(ctx context.Context, accountID int, id int) error {
	delete(r.byID, id)
	return nil
}

type fakeRoomServiceUnitRepo struct {
	byID map[int]*domainunit.Unit
}

func (r *fakeRoomServiceUnitRepo) Create(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeRoomServiceUnitRepo) Get(ctx context.Context, accountID int, id int) (*domainunit.Unit, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeRoomServiceUnitRepo) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	unit := r.byID[id]
	if unit == nil {
		return nil, domainunit.ErrNotFound
	}
	return cloneUnit(unit), nil
}

func (r *fakeRoomServiceUnitRepo) List(ctx context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	units := make([]*domainunit.Unit, 0, len(r.byID))
	for _, unit := range r.byID {
		units = append(units, cloneUnit(unit))
	}
	return units, nil
}

func (r *fakeRoomServiceUnitRepo) Update(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	return unit, nil
}

func (r *fakeRoomServiceUnitRepo) Delete(ctx context.Context, accountID int, id int) error {
	delete(r.byID, id)
	return nil
}

func cloneRoom(room *domainroom.Room) *domainroom.Room {
	if room == nil {
		return nil
	}
	cloned := *room
	cloned.MeterItemIDs = append([]int(nil), room.MeterItemIDs...)
	return &cloned
}

func cloneUnit(unit *domainunit.Unit) *domainunit.Unit {
	if unit == nil {
		return nil
	}
	cloned := *unit
	cloned.DefaultMeterItemIDs = append([]int(nil), unit.DefaultMeterItemIDs...)
	return &cloned
}

var _ domainroom.Repository = (*fakeRoomServiceRoomRepo)(nil)
var _ domainunit.Repository = (*fakeRoomServiceUnitRepo)(nil)
