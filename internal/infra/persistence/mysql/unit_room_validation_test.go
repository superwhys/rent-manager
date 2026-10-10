package mysql

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestUnitRepositoryByteLengthValidation(t *testing.T) {
	for _, field := range []string{"name", "address", "city", "province"} {
		t.Run(field, func(t *testing.T) {
			ctx := context.Background()
			db := newTestDB(t)
			q := query.Use(db)
			owner := testAccountID("owner")
			water := createUnitTestMeterItem(t, q, owner, "water")
			power := createUnitTestMeterItem(t, q, owner, "power")
			repo := NewUnitRepository(db)
			unit := &domainunit.Unit{
				AccountID:           owner,
				Name:                "name",
				Address:             "address",
				City:                "city",
				Province:            "province",
				Config:              domainunit.DefaultUnitConfig(),
				DefaultMeterItemIDs: []int{water.ID},
			}
			setField := func(unit *domainunit.Unit, value string) {
				switch field {
				case "name":
					unit.Name = value
				case "address":
					unit.Address = value
				case "city":
					unit.City = value
				case "province":
					unit.Province = value
				}
			}
			boundary := strings.Repeat("房", 85)
			setField(unit, boundary)
			stored, err := repo.Create(ctx, unit)
			if err != nil {
				t.Fatalf("255-byte Create: %v", err)
			}
			if _, err := repo.Update(ctx, stored); err != nil {
				t.Fatalf("255-byte Update: %v", err)
			}

			invalid := *stored
			setField(&invalid, boundary+"a")
			invalid.Config.MeterReadingDay = 31
			invalid.DefaultMeterItemIDs = []int{power.ID}
			if _, err := repo.Create(ctx, &invalid); err == nil {
				t.Fatal("256-byte Create succeeded")
			} else if errors.Is(err, domainunit.ErrNameAlreadyExists) {
				t.Fatalf("256-byte Create reached the database: %v", err)
			}
			if _, err := repo.Update(ctx, &invalid); err == nil {
				t.Fatal("256-byte Update succeeded")
			}
			count, err := q.Unit.WithContext(ctx).Count()
			if err != nil || count != 1 {
				t.Fatalf("invalid Create wrote a unit: count=%d, err=%v", count, err)
			}
			after, err := repo.Get(ctx, owner, stored.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.Name != stored.Name || after.Address != stored.Address ||
				after.City != stored.City ||
				after.Province != stored.Province ||
				after.Config != stored.Config ||
				!reflect.DeepEqual(after.DefaultMeterItemIDs, stored.DefaultMeterItemIDs) {
				t.Fatalf("invalid Update changed the unit: %+v", after)
			}
			invalid.DefaultMeterItemIDs = []int{power.ID + 1}
			if _, err := repo.Create(
				ctx,
				&invalid,
			); !errors.Is(
				err,
				domainunit.ErrMeterItemNotFound,
			) {
				t.Fatalf("relation validation precedence: %v", err)
			}
		})
	}
}

func TestRoomRepositoryByteLengthValidation(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	water := createUnitTestMeterItem(t, q, owner, "water")
	power := createUnitTestMeterItem(t, q, owner, "power")
	room := createBillingTestRoom(t, db, q, owner, "owned", []int{water.ID})
	repo := NewRoomRepository(db)
	room.RoomNo = strings.Repeat("房", 85)
	stored, err := repo.Update(ctx, owner, room)
	if err != nil {
		t.Fatalf("255-byte Update: %v", err)
	}
	if _, err := repo.Create(ctx, owner, room); err != nil {
		t.Fatalf("255-byte Create: %v", err)
	}
	invalid := *stored
	invalid.RoomNo += "a"
	invalid.Price = 999
	invalid.MeterItemIDs = []int{power.ID}
	if _, err := repo.Create(ctx, owner, &invalid); err == nil {
		t.Fatal("256-byte Create succeeded")
	}
	if _, err := repo.Update(ctx, owner, &invalid); err == nil {
		t.Fatal("256-byte Update succeeded")
	}
	count, err := q.Room.WithContext(ctx).Count()
	if err != nil || count != 2 {
		t.Fatalf("invalid Create wrote a room: count=%d, err=%v", count, err)
	}
	after, err := repo.Get(ctx, owner, stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.RoomNo != stored.RoomNo || after.Price != stored.Price ||
		!reflect.DeepEqual(after.MeterItemIDs, stored.MeterItemIDs) {
		t.Fatalf("invalid Update changed the room: %+v", after)
	}
	invalid.UnitID++
	if _, err := repo.Create(ctx, owner, &invalid); !errors.Is(err, domainroom.ErrUnitRequired) {
		t.Fatalf("relation validation precedence: %v", err)
	}
}
