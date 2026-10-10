package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestRentBillRepositoryOwnershipNullableValuesAndFilters(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	other := testAccountID("other")
	room := createBillingTestRoom(t, db, q, owner, "owned", nil)
	otherRoom := createBillingTestRoom(t, db, q, other, "other", nil)
	now, remark := time.Now(), "paid"
	repo := NewRentBillRepository(db)
	bill, err := repo.Create(ctx, owner, &domainrentbill.RentBill{
		RoomID: room.ID, PeriodMonth: "2026-09", BaseRent: 100, MeterFee: 20, TotalAmount: 120,
		Status: domainrentbill.StatusPaid, PaidAt: &now, Remark: &remark,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, owner, bill); !errors.Is(err, domainrentbill.ErrAlreadyExists) {
		t.Fatalf("duplicate Create error = %v", err)
	}
	if _, err := repo.Get(ctx, other, bill.ID); !errors.Is(err, domainrentbill.ErrNotFound) {
		t.Fatalf("foreign Get error = %v", err)
	}
	if _, err := repo.GetByRoomPeriod(
		ctx,
		other,
		room.ID,
		bill.PeriodMonth,
	); !errors.Is(
		err,
		domainrentbill.ErrNotFound,
	) {
		t.Fatalf("foreign period Get error = %v", err)
	}
	if _, err := repo.Update(ctx, owner, bill); err != nil {
		t.Fatalf("unchanged update: %v", err)
	}
	bill.BaseRent, bill.MeterFee, bill.TotalAmount = 0, 0, 0
	bill.Status = domainrentbill.StatusUnpaid
	bill.PaidAt, bill.Remark = nil, nil
	bill, err = repo.Update(ctx, owner, bill)
	if err != nil {
		t.Fatal(err)
	}
	if bill.BaseRent != 0 || bill.MeterFee != 0 || bill.TotalAmount != 0 ||
		bill.Status != domainrentbill.StatusUnpaid ||
		bill.PaidAt != nil ||
		bill.Remark != nil {
		t.Fatalf("zero/nullable values not saved: %+v", bill)
	}
	later := *bill
	later.ID = 0
	later.PeriodMonth = "2026-10"
	created, err := repo.Create(ctx, owner, &later)
	if err != nil {
		t.Fatal(err)
	}
	items, err := repo.List(ctx, owner, domainrentbill.ListQuery{
		UnitID: room.UnitID, RoomID: room.ID, PeriodMonthFrom: "2026-09", PeriodMonthTo: "2026-10",
		Status: domainrentbill.StatusUnpaid, Limit: 1, Offset: 1,
	})
	if err != nil || len(items) != 1 || items[0].ID != bill.ID {
		t.Fatalf("ordered bill page = %+v, %v", items, err)
	}
	items, err = repo.List(ctx, owner, domainrentbill.ListQuery{PeriodMonth: "2026-10"})
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("period filtered bills = %+v, %v", items, err)
	}
	items, err = repo.List(ctx, other, domainrentbill.ListQuery{})
	if err != nil || len(items) != 0 {
		t.Fatalf("foreign bills = %+v, %v", items, err)
	}
	foreignUpdate := *bill
	foreignUpdate.RoomID = otherRoom.ID
	if _, err := repo.Update(
		ctx,
		other,
		&foreignUpdate,
	); !errors.Is(
		err,
		domainrentbill.ErrNotFound,
	) {
		t.Fatalf("foreign Update error = %v", err)
	}
	if err := repo.Delete(ctx, other, bill.ID); !errors.Is(err, domainrentbill.ErrNotFound) {
		t.Fatalf("foreign Delete error = %v", err)
	}
	if err := repo.Delete(ctx, owner, bill.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, bill.ID); !errors.Is(err, domainrentbill.ErrNotFound) {
		t.Fatalf("deleted GetByID error = %v", err)
	}
	if _, err := repo.GetByRoomPeriod(
		ctx,
		owner,
		room.ID,
		bill.PeriodMonth,
	); !errors.Is(
		err,
		domainrentbill.ErrNotFound,
	) {
		t.Fatalf("deleted period Get error = %v", err)
	}
	if _, err := q.Room.WithContext(ctx).
		Where(q.Room.ID.Eq(room.ID)).
		UpdateSimple(q.Room.DeletedAt.Value(time.Now())); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, owner, created.ID); !errors.Is(err, domainrentbill.ErrNotFound) {
		t.Fatalf("bill in deleted room Get error = %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); err != nil {
		t.Fatalf("unscoped bill GetByID in deleted room: %v", err)
	}
}

func TestRentBillRepositorySaveGeneratedBillsTransaction(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	room := createBillingTestRoom(t, db, q, owner, "owned", nil)
	repo := NewRentBillRepository(db)
	now := time.Now()
	unpaid := &domainrentbill.RentBill{
		RoomID:      room.ID,
		PeriodMonth: "2026-07",
		BaseRent:    100,
		TotalAmount: 100,
		Status:      domainrentbill.StatusUnpaid,
	}
	paid := &domainrentbill.RentBill{
		RoomID:      room.ID,
		PeriodMonth: "2026-08",
		BaseRent:    200,
		TotalAmount: 200,
		Status:      domainrentbill.StatusPaid,
		PaidAt:      &now,
	}
	cancelled := &domainrentbill.RentBill{
		RoomID:      room.ID,
		PeriodMonth: "2026-09",
		BaseRent:    300,
		TotalAmount: 300,
		Status:      domainrentbill.StatusCancelled,
	}
	for _, bill := range []*domainrentbill.RentBill{unpaid, paid, cancelled} {
		saved, err := repo.Create(ctx, owner, bill)
		if err != nil {
			t.Fatal(err)
		}
		bill.ID = saved.ID
	}
	generated := make([]*domainrentbill.RentBill, 0, 4)
	for _, month := range []string{"2026-07", "2026-08", "2026-09", "2026-10"} {
		generated = append(generated, &domainrentbill.RentBill{
			RoomID:      room.ID,
			PeriodMonth: month,
			BaseRent:    400,
			MeterFee:    50,
			TotalAmount: 450,
			Status:      domainrentbill.StatusUnpaid,
		})
	}
	result, err := repo.SaveGeneratedBills(ctx, owner, generated)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Saved) != 2 || len(result.Skipped) != 2 || result.Saved[0].ID != unpaid.ID ||
		result.Saved[0].TotalAmount != 450 ||
		result.Saved[1].PeriodMonth != "2026-10" {
		t.Fatalf("unexpected generated result: %+v", result)
	}
	if result.Skipped[0].ID != paid.ID || result.Skipped[0].Status != domainrentbill.StatusPaid ||
		result.Skipped[0].TotalAmount != 200 ||
		result.Skipped[1].ID != cancelled.ID ||
		result.Skipped[1].Status != domainrentbill.StatusCancelled ||
		result.Skipped[1].TotalAmount != 300 {
		t.Fatalf("paid/cancelled bills changed: %+v", result.Skipped)
	}

	writeErr := errors.New("bill write failed")
	if err := db.Callback().
		Create().
		Before("gorm:create").
		Register("rent_bill_test:fail_create", func(tx *gorm.DB) {
			if tx.Statement.Table == "rent_bills" {
				tx.AddError(writeErr)
			}
		}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Callback().Create().Remove("rent_bill_test:fail_create") })
	updated := *result.Saved[0]
	updated.BaseRent, updated.TotalAmount = 999, 1049
	added := updated
	added.ID = 0
	added.PeriodMonth = "2026-11"
	if _, err := repo.SaveGeneratedBills(
		ctx,
		owner,
		[]*domainrentbill.RentBill{&updated, &added},
	); !errors.Is(
		err,
		writeErr,
	) {
		t.Fatalf("failed generated save error = %v", err)
	}
	stored, err := repo.Get(ctx, owner, unpaid.ID)
	if err != nil || stored.TotalAmount != 450 {
		t.Fatalf("generated update rollback = %+v, %v", stored, err)
	}
	if _, err := repo.GetByRoomPeriod(
		ctx,
		owner,
		room.ID,
		"2026-11",
	); !errors.Is(
		err,
		domainrentbill.ErrNotFound,
	) {
		t.Fatalf("generated create rollback error = %v", err)
	}
	if _, err := repo.SaveGeneratedBills(
		ctx,
		owner,
		nil,
	); !errors.Is(
		err,
		domainrentbill.ErrBillsRequired,
	) {
		t.Fatalf("empty generated save error = %v", err)
	}
}
