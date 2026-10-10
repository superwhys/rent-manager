package mysql

import (
	"context"
	"reflect"
	"testing"
	"time"

	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

func TestRentBillRepositoryMeterReadingSnapshotsRoundTrip(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	room := createBillingTestRoom(t, db, q, owner, "owned", nil)
	repo := NewRentBillRepository(db)
	want := rentBillTestSnapshots(room.ID, "2026-10", owner)
	bill, err := repo.Create(ctx, owner, &domainrentbill.RentBill{
		RoomID: room.ID, PeriodMonth: "2026-10", Status: domainrentbill.StatusUnpaid,
		MeterFee: 8, TotalAmount: 8, MeterReadings: want,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRentBillSnapshots(t, want, bill.MeterReadings)
	for _, get := range []struct {
		name string
		load func() (*domainrentbill.RentBill, error)
	}{
		{name: "Get", load: func() (*domainrentbill.RentBill, error) {
			return repo.Get(ctx, owner, bill.ID)
		}},
		{name: "GetByID", load: func() (*domainrentbill.RentBill, error) {
			return repo.GetByID(ctx, bill.ID)
		}},
		{name: "GetByRoomPeriod", load: func() (*domainrentbill.RentBill, error) {
			return repo.GetByRoomPeriod(ctx, owner, room.ID, bill.PeriodMonth)
		}},
	} {
		t.Run(get.name, func(t *testing.T) {
			stored, err := get.load()
			if err != nil {
				t.Fatal(err)
			}
			assertRentBillSnapshots(t, want, stored.MeterReadings)
		})
	}
	items, err := repo.List(ctx, owner, domainrentbill.ListQuery{})
	if err != nil || len(items) != 1 {
		t.Fatalf("List = %+v, %v", items, err)
	}
	assertRentBillSnapshots(t, want, items[0].MeterReadings)
}

func TestRentBillRepositoryMeterReadingSnapshotsNullAndEmpty(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	room := createBillingTestRoom(t, db, q, owner, "owned", nil)
	repo := NewRentBillRepository(db)
	for _, scenario := range []struct {
		name     string
		month    string
		readings []domainrentbill.MeterReadingSnapshot
	}{
		{name: "历史账单无快照", month: "2026-09", readings: nil},
		{name: "已生成空快照", month: "2026-10", readings: []domainrentbill.MeterReadingSnapshot{}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			bill, err := repo.Create(ctx, owner, &domainrentbill.RentBill{
				RoomID: room.ID, PeriodMonth: scenario.month, Status: domainrentbill.StatusUnpaid,
				MeterReadings: scenario.readings,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertRentBillSnapshots(t, scenario.readings, bill.MeterReadings)
			record, err := q.RentBill.WithContext(ctx).Where(q.RentBill.ID.Eq(bill.ID)).Take()
			if err != nil {
				t.Fatal(err)
			}
			if scenario.readings == nil {
				if record.MeterReadingsJSON != nil {
					t.Fatalf("无快照应保存 SQL NULL，实际为 %q", *record.MeterReadingsJSON)
				}
			} else if record.MeterReadingsJSON == nil || *record.MeterReadingsJSON != "[]" {
				t.Fatalf("空快照应保存 []，实际为 %+v", record.MeterReadingsJSON)
			}
			bill.MeterReadings = rentBillTestSnapshots(room.ID, scenario.month, owner)
			if _, err := repo.Update(ctx, owner, bill); err != nil {
				t.Fatal(err)
			}
			bill.MeterReadings = scenario.readings
			stored, err := repo.Update(ctx, owner, bill)
			if err != nil {
				t.Fatal(err)
			}
			assertRentBillSnapshots(t, scenario.readings, stored.MeterReadings)
			record, err = q.RentBill.WithContext(ctx).Where(q.RentBill.ID.Eq(bill.ID)).Take()
			if err != nil {
				t.Fatal(err)
			}
			if (record.MeterReadingsJSON == nil) != (scenario.readings == nil) {
				t.Fatalf("Update 未保存 NULL/[] 区别: %+v", record.MeterReadingsJSON)
			}
		})
	}
}

func TestRentBillRepositoryUpdatesPreserveMeterReadingSnapshots(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*domainrentbill.RentBill) error
	}{
		{name: "MarkPaid", mutate: func(bill *domainrentbill.RentBill) error {
			return bill.MarkPaid(time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC))
		}},
		{name: "Cancel", mutate: func(bill *domainrentbill.RentBill) error {
			return bill.Cancel()
		}},
		{name: "Update", mutate: func(bill *domainrentbill.RentBill) error {
			return bill.Update(domainrentbill.UpdateParams{
				ID: bill.ID, RoomID: bill.RoomID, PeriodMonth: bill.PeriodMonth,
				BaseRent: 500, MeterFee: bill.MeterFee, Status: bill.Status,
			})
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := context.Background()
			db := newTestDB(t)
			q := query.Use(db)
			owner := testAccountID("owner")
			room := createBillingTestRoom(t, db, q, owner, "owned", nil)
			repo := NewRentBillRepository(db)
			want := rentBillTestSnapshots(room.ID, "2026-10", owner)
			bill, err := repo.Create(ctx, owner, &domainrentbill.RentBill{
				RoomID: room.ID, PeriodMonth: "2026-10", Status: domainrentbill.StatusUnpaid,
				MeterFee: 8, TotalAmount: 8, MeterReadings: want,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := scenario.mutate(bill); err != nil {
				t.Fatal(err)
			}
			stored, err := repo.Update(ctx, owner, bill)
			if err != nil {
				t.Fatal(err)
			}
			assertRentBillSnapshots(t, want, stored.MeterReadings)
		})
	}
}

func TestRentBillRepositoryRejectsInvalidMeterReadingSnapshots(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	q := query.Use(db)
	owner := testAccountID("owner")
	room := createBillingTestRoom(t, db, q, owner, "owned", nil)
	repo := NewRentBillRepository(db)
	bill, err := repo.Create(ctx, owner, &domainrentbill.RentBill{
		RoomID: room.ID, PeriodMonth: "2026-10", Status: domainrentbill.StatusUnpaid,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.RentBill.WithContext(ctx).Where(q.RentBill.ID.Eq(bill.ID)).
		Update(q.RentBill.MeterReadingsJSON, `{"broken":"快照格式不正确"}`); err != nil {
		t.Fatal(err)
	}
	for _, get := range []struct {
		name string
		load func() error
	}{
		{name: "Get", load: func() error {
			_, err := repo.Get(ctx, owner, bill.ID)
			return err
		}},
		{name: "GetByID", load: func() error {
			_, err := repo.GetByID(ctx, bill.ID)
			return err
		}},
		{name: "GetByRoomPeriod", load: func() error {
			_, err := repo.GetByRoomPeriod(ctx, owner, room.ID, bill.PeriodMonth)
			return err
		}},
		{name: "List", load: func() error {
			_, err := repo.List(ctx, owner, domainrentbill.ListQuery{})
			return err
		}},
	} {
		t.Run(get.name, func(t *testing.T) {
			if err := get.load(); err == nil {
				t.Fatal("损坏快照应返回错误")
			}
		})
	}
}

func rentBillTestSnapshots(roomID int, month, accountID string) []domainrentbill.MeterReadingSnapshot {
	zero, empty, itemRemark := 0.0, "", "阶梯水价说明"
	readingTime := time.Date(2026, 10, 1, 8, 30, 0, 123456000, time.UTC)
	return []domainrentbill.MeterReadingSnapshot{
		{
			ID: 11, RoomID: roomID, PeriodMonth: month,
			MeterItem: &domainmeteritem.MeterItem{
				ID: 21, AccountID: accountID, Code: "水表", Unit: "吨", UnitPrice: 3.1415,
				IsSystem: true, CalcType: domainmeteritem.CalcTypeDifference, Remark: &itemRemark,
				CreatedAt: readingTime, UpdatedAt: readingTime,
			},
			PreviousReading: &zero, CurrentReading: 0, UsageAmount: &zero,
			ReadingTime: readingTime, ReaderID: &accountID, Source: domainmeterreading.SourceManual,
			Remark: &empty, Fee: 0, IsRecorded: true,
		},
		{
			ID: 12, RoomID: roomID, PeriodMonth: month,
			MeterItem: &domainmeteritem.MeterItem{
				ID: 22, AccountID: accountID, Code: "电费", Unit: "度", UnitPrice: 0.625,
				CalcType: domainmeteritem.CalcTypeDirect,
			},
			CurrentReading: 12.75, ReadingTime: readingTime, Source: domainmeterreading.SourceImport,
			Fee: 8, IsRecorded: true,
		},
		{
			RoomID: roomID, PeriodMonth: month,
			MeterItem: &domainmeteritem.MeterItem{
				ID: 23, AccountID: accountID, Code: "燃气表", Unit: "立方米", UnitPrice: 2.75,
				CalcType: domainmeteritem.CalcTypeDifference,
			},
			IsRecorded: false,
		},
	}
}

func assertRentBillSnapshots(t *testing.T, want, got []domainrentbill.MeterReadingSnapshot) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("账单抄表快照未完整保留:\nwant: %+v\ngot:  %+v", want, got)
	}
}
