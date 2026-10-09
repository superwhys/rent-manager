package mysql

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"

	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestRoomRepositoryOwnershipFiltersAndZeroValues(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := createUnitTestAccount(t, q, "owner")
	other := createUnitTestAccount(t, q, "other")
	unit := createUnitTestUnit(t, q, owner.ID, "owned")
	otherUnit := createUnitTestUnit(t, q, other.ID, "other")
	roomType := &models.RoomType{AccountID: owner.ID, Name: "standard"}
	otherRoomType := &models.RoomType{AccountID: other.ID, Name: "standard"}
	if err := q.RoomType.WithContext(ctx).Create(roomType, otherRoomType); err != nil {
		t.Fatal(err)
	}
	water := createUnitTestMeterItem(t, q, owner.ID, "water")
	power := createUnitTestMeterItem(t, q, owner.ID, "power")
	foreign := createUnitTestMeterItem(t, q, other.ID, "water")
	repo := NewRoomRepository(db)
	room, err := repo.Create(ctx, owner.ID, &domainroom.Room{
		RoomNo: "101", Floor: 1, Price: 100, Size: 20, UnitID: unit.ID, RoomTypeID: roomType.ID,
		IsActive: false, MeterItemIDs: []int{water.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if room.IsActive || !reflect.DeepEqual(room.MeterItemIDs, []int{water.ID}) {
		t.Fatalf("unexpected room: %+v", room)
	}
	second, err := repo.Create(ctx, owner.ID, &domainroom.Room{
		RoomNo: "102", Floor: 1, Price: 100, Size: 20, UnitID: unit.ID, RoomTypeID: roomType.ID, IsActive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, other.ID, room.ID); !errors.Is(err, domainroom.ErrNotFound) {
		t.Fatalf("foreign Get error = %v", err)
	}
	rooms, err := repo.List(ctx, owner.ID, domainroom.ListQuery{UnitID: unit.ID, Limit: 1})
	if err != nil || len(rooms) != 1 || rooms[0].ID != second.ID {
		t.Fatalf("ordered page = %+v, %v", rooms, err)
	}
	rooms, err = repo.List(ctx, owner.ID, domainroom.ListQuery{UnitID: unit.ID, Limit: 1, Offset: 1})
	if err != nil || len(rooms) != 1 || rooms[0].ID != room.ID {
		t.Fatalf("offset page = %+v, %v", rooms, err)
	}
	inactive := false
	count, err := repo.Count(ctx, owner.ID, domainroom.ListQuery{UnitID: unit.ID, RoomTypeID: roomType.ID, IsActive: &inactive, Offset: 9, Limit: 1})
	if err != nil || count != 1 {
		t.Fatalf("filtered count = %d, %v", count, err)
	}
	count, err = repo.Count(ctx, other.ID, domainroom.ListQuery{UnitID: unit.ID})
	if err != nil || count != 0 {
		t.Fatalf("foreign count = %d, %v", count, err)
	}
	rooms, err = repo.List(ctx, other.ID, domainroom.ListQuery{UnitID: unit.ID})
	if err != nil || len(rooms) != 0 {
		t.Fatalf("foreign list = %+v, %v", rooms, err)
	}
	if _, err := repo.Update(ctx, owner.ID, room); err != nil {
		t.Fatalf("unchanged update: %v", err)
	}
	room.Floor = 0
	room.Price = 0
	room.MeterItemIDs = []int{power.ID}
	room, err = repo.Update(ctx, owner.ID, room)
	if err != nil {
		t.Fatal(err)
	}
	if room.Floor != 0 || room.Price != 0 || room.IsActive || !reflect.DeepEqual(room.MeterItemIDs, []int{power.ID}) {
		t.Fatalf("zero values/relations not saved: %+v", room)
	}

	for _, test := range []struct {
		name   string
		change func(*domainroom.Room)
		want   error
	}{
		{"foreign unit", func(r *domainroom.Room) { r.UnitID = otherUnit.ID }, domainroom.ErrUnitRequired},
		{"foreign type", func(r *domainroom.Room) { r.RoomTypeID = otherRoomType.ID }, domainroom.ErrRoomTypeRequired},
		{"foreign item", func(r *domainroom.Room) { r.MeterItemIDs = []int{foreign.ID} }, domainroom.ErrMeterItemNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := *room
			test.change(&invalid)
			if _, err := repo.Create(ctx, owner.ID, &invalid); !errors.Is(err, test.want) {
				t.Fatalf("Create error = %v, want %v", err, test.want)
			}
			if _, err := repo.Update(ctx, owner.ID, &invalid); !errors.Is(err, test.want) {
				t.Fatalf("Update error = %v, want %v", err, test.want)
			}
		})
	}
	foreignUpdate := *room
	foreignUpdate.UnitID = otherUnit.ID
	foreignUpdate.RoomTypeID = otherRoomType.ID
	foreignUpdate.MeterItemIDs = nil
	if _, err := repo.Update(ctx, other.ID, &foreignUpdate); !errors.Is(err, domainroom.ErrNotFound) {
		t.Fatalf("foreign Update error = %v", err)
	}
	if err := repo.Delete(ctx, other.ID, room.ID); !errors.Is(err, domainroom.ErrNotFound) {
		t.Fatalf("foreign Delete error = %v", err)
	}
	if err := NewUnitRepository(db).Delete(ctx, owner.ID, unit.ID); !errors.Is(err, domainunit.ErrUnitHasRooms) {
		t.Fatalf("occupied unit Delete error = %v", err)
	}
	room.MeterItemIDs = nil
	room, err = repo.Update(ctx, owner.ID, room)
	if err != nil || len(room.MeterItemIDs) != 0 {
		t.Fatalf("clear meter items = %+v, %v", room, err)
	}
	if err := repo.Delete(ctx, owner.ID, room.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, room.ID); !errors.Is(err, domainroom.ErrNotFound) {
		t.Fatalf("deleted GetByID error = %v", err)
	}
	stored, err := q.Room.WithContext(ctx).Where(q.Room.ID.Eq(room.ID)).Take()
	if err != nil || stored.DeletedAt == nil || *stored.IsActive {
		t.Fatalf("soft deleted room = %+v, %v", stored, err)
	}
	if _, err := q.Unit.WithContext(ctx).Where(q.Unit.ID.Eq(unit.ID)).UpdateSimple(q.Unit.DeletedAt.Value(time.Now())); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, owner.ID, second.ID); !errors.Is(err, domainroom.ErrNotFound) {
		t.Fatalf("room in deleted unit Get error = %v", err)
	}
	if _, err := repo.GetByID(ctx, second.ID); err != nil {
		t.Fatalf("unscoped GetByID in deleted unit: %v", err)
	}
}

func TestRoomRepositoryMeterItemTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := createUnitTestAccount(t, q, "owner")
	unit := createUnitTestUnit(t, q, owner.ID, "owned")
	roomType := &models.RoomType{AccountID: owner.ID, Name: "standard"}
	if err := q.RoomType.WithContext(ctx).Create(roomType); err != nil {
		t.Fatal(err)
	}
	water := createUnitTestMeterItem(t, q, owner.ID, "water")
	power := createUnitTestMeterItem(t, q, owner.ID, "power")
	repo := NewRoomRepository(db)
	room, err := repo.Create(ctx, owner.ID, &domainroom.Room{
		RoomNo: "before", Floor: 1, Price: 100, Size: 20, UnitID: unit.ID, RoomTypeID: roomType.ID,
		IsActive: true, MeterItemIDs: []int{water.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	writeErr := errors.New("relation write failed")
	if err := db.Callback().Create().Before("gorm:create").Register("room_test:fail_relation", func(tx *gorm.DB) {
		if tx.Statement.Table == "room_meter_items" {
			tx.AddError(writeErr)
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove("room_test:fail_relation") })

	room.RoomNo = "after"
	room.MeterItemIDs = []int{power.ID}
	if _, err := repo.Update(ctx, owner.ID, room); !errors.Is(err, writeErr) {
		t.Fatalf("Update error = %v", err)
	}
	stored, err := repo.Get(ctx, owner.ID, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.RoomNo != "before" || !reflect.DeepEqual(stored.MeterItemIDs, []int{water.ID}) {
		t.Fatalf("failed update was not rolled back: %+v", stored)
	}
	if _, err := repo.Create(ctx, owner.ID, room); !errors.Is(err, writeErr) {
		t.Fatalf("Create error = %v", err)
	}
	count, err := q.Room.WithContext(ctx).Count()
	if err != nil || count != 1 {
		t.Fatalf("failed create was not rolled back: count=%d, err=%v", count, err)
	}
}
