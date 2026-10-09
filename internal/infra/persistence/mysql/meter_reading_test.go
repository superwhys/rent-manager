package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestMeterReadingRepositoryOwnershipNullableValuesAndFilters(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := createUnitTestAccount(t, q, "owner")
	other := createUnitTestAccount(t, q, "other")
	water := createUnitTestMeterItem(t, q, owner.ID, "water")
	foreign := createUnitTestMeterItem(t, q, other.ID, "water")
	room := createBillingTestRoom(t, db, q, owner.ID, "owned", []int{water.ID})
	otherRoom := createBillingTestRoom(t, db, q, other.ID, "other", []int{foreign.ID})
	previous, usage, reader, remark := 5.0, 5.0, int64(owner.ID), "read"
	repo := NewMeterReadingRepository(db)
	reading, err := repo.Create(ctx, owner.ID, &domainmeterreading.MeterReading{
		RoomID: room.ID, MeterItemID: water.ID, PeriodMonth: "2026-09", PreviousReading: &previous,
		CurrentReading: 10, UsageAmount: &usage, ReadingTime: time.Now(), ReaderID: &reader,
		Source: domainmeterreading.SourceManual, Remark: &remark,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, owner.ID, reading); !errors.Is(err, domainmeterreading.ErrAlreadyExists) {
		t.Fatalf("duplicate Create error = %v", err)
	}
	if _, err := repo.Get(ctx, other.ID, reading.ID); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("foreign Get error = %v", err)
	}
	if _, err := repo.GetByRoomMeterItemPeriod(ctx, other.ID, room.ID, water.ID, reading.PeriodMonth); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("foreign period Get error = %v", err)
	}
	if _, err := repo.Update(ctx, owner.ID, reading); err != nil {
		t.Fatalf("unchanged update: %v", err)
	}
	zero, zeroReader, empty := 0.0, int64(0), ""
	reading.PreviousReading = &zero
	reading.CurrentReading = 0
	reading.UsageAmount = &zero
	reading.ReaderID = &zeroReader
	reading.Remark = &empty
	reading, err = repo.Update(ctx, owner.ID, reading)
	if err != nil {
		t.Fatal(err)
	}
	if reading.PreviousReading == nil || *reading.PreviousReading != 0 || reading.CurrentReading != 0 || reading.UsageAmount == nil || *reading.UsageAmount != 0 || reading.ReaderID == nil || *reading.ReaderID != 0 || reading.Remark == nil || *reading.Remark != "" {
		t.Fatalf("zero values not saved: %+v", reading)
	}
	reading.PreviousReading, reading.UsageAmount, reading.ReaderID, reading.Remark = nil, nil, nil, nil
	reading, err = repo.Update(ctx, owner.ID, reading)
	if err != nil {
		t.Fatal(err)
	}
	if reading.PreviousReading != nil || reading.UsageAmount != nil || reading.ReaderID != nil || reading.Remark != nil {
		t.Fatalf("nullable fields not cleared: %+v", reading)
	}
	later := *reading
	later.ID = 0
	later.PeriodMonth = "2026-10"
	later.Source = domainmeterreading.SourceImport
	created, err := repo.Create(ctx, owner.ID, &later)
	if err != nil {
		t.Fatal(err)
	}
	latest, err := repo.GetLatestBeforePeriod(ctx, owner.ID, room.ID, water.ID, "2026-11")
	if err != nil || latest.ID != created.ID {
		t.Fatalf("latest reading = %+v, %v", latest, err)
	}
	items, err := repo.List(ctx, owner.ID, domainmeterreading.ListQuery{
		UnitID: room.UnitID, RoomID: room.ID, MeterItemID: water.ID, PeriodMonthFrom: "2026-09", PeriodMonthTo: "2026-10", Limit: 1, Offset: 1,
	})
	if err != nil || len(items) != 1 || items[0].ID != reading.ID {
		t.Fatalf("ordered reading page = %+v, %v", items, err)
	}
	items, err = repo.List(ctx, owner.ID, domainmeterreading.ListQuery{PeriodMonth: "2026-10", Source: domainmeterreading.SourceImport})
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("filtered readings = %+v, %v", items, err)
	}
	items, err = repo.List(ctx, other.ID, domainmeterreading.ListQuery{})
	if err != nil || len(items) != 0 {
		t.Fatalf("foreign readings = %+v, %v", items, err)
	}
	foreignUpdate := *reading
	foreignUpdate.RoomID, foreignUpdate.MeterItemID = otherRoom.ID, foreign.ID
	if _, err := repo.Update(ctx, other.ID, &foreignUpdate); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("foreign Update error = %v", err)
	}
	if err := repo.Delete(ctx, other.ID, reading.ID); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("foreign Delete error = %v", err)
	}
	if err := repo.Delete(ctx, owner.ID, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, owner.ID, created.ID); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("deleted Get error = %v", err)
	}
	latest, err = repo.GetLatestBeforePeriod(ctx, owner.ID, room.ID, water.ID, "2026-11")
	if err != nil || latest.ID != reading.ID {
		t.Fatalf("latest excludes deleted = %+v, %v", latest, err)
	}
}

func TestMeterReadingRepositorySaveRoomMonthTransaction(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := createUnitTestAccount(t, q, "owner")
	water := createUnitTestMeterItem(t, q, owner.ID, "water")
	power := createUnitTestMeterItem(t, q, owner.ID, "power")
	unassigned := createUnitTestMeterItem(t, q, owner.ID, "unassigned")
	room := createBillingTestRoom(t, db, q, owner.ID, "owned", []int{water.ID, power.ID})
	repo := NewMeterReadingRepository(db)
	reading, err := repo.Create(ctx, owner.ID, &domainmeterreading.MeterReading{
		RoomID: room.ID, MeterItemID: water.ID, PeriodMonth: "2026-10", CurrentReading: 10,
		ReadingTime: time.Now(), Source: domainmeterreading.SourceManual,
	})
	if err != nil {
		t.Fatal(err)
	}
	updated := *reading
	updated.CurrentReading = 20
	added := updated
	added.ID = 0
	added.MeterItemID = power.ID
	invalid := added
	invalid.MeterItemID = unassigned.ID
	if _, err := repo.SaveRoomMonth(ctx, owner.ID, []*domainmeterreading.MeterReading{&updated, &added, &invalid}); !errors.Is(err, domainmeterreading.ErrRoomMeterItemNotAssigned) {
		t.Fatalf("SaveRoomMonth error = %v", err)
	}
	stored, err := repo.Get(ctx, owner.ID, reading.ID)
	if err != nil || stored.CurrentReading != 10 {
		t.Fatalf("update rollback = %+v, %v", stored, err)
	}
	if _, err := repo.GetByRoomMeterItemPeriod(ctx, owner.ID, room.ID, power.ID, "2026-10"); !errors.Is(err, domainmeterreading.ErrNotFound) {
		t.Fatalf("create rollback error = %v", err)
	}
	saved, err := repo.SaveRoomMonth(ctx, owner.ID, []*domainmeterreading.MeterReading{&updated, &added})
	if err != nil || len(saved) != 2 || saved[0].ID != reading.ID || saved[0].CurrentReading != 20 || saved[1].MeterItemID != power.ID {
		t.Fatalf("saved readings = %+v, %v", saved, err)
	}
	if _, err := repo.SaveRoomMonth(ctx, owner.ID, nil); !errors.Is(err, domainmeterreading.ErrReadingsRequired) {
		t.Fatalf("empty save error = %v", err)
	}
}

func createBillingTestRoom(t *testing.T, db *gorm.DB, q *query.Query, accountID int, name string, meterItemIDs []int) *domainroom.Room {
	t.Helper()
	unit := createUnitTestUnit(t, q, accountID, name)
	roomType := &models.RoomType{AccountID: accountID, Name: name}
	if err := q.RoomType.WithContext(context.Background()).Create(roomType); err != nil {
		t.Fatal(err)
	}
	room, err := NewRoomRepository(db).Create(context.Background(), accountID, &domainroom.Room{
		RoomNo: "101", Floor: 1, Price: 100, Size: 20, UnitID: unit.ID, RoomTypeID: roomType.ID,
		IsActive: true, MeterItemIDs: meterItemIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	return room
}
