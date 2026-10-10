package rentbill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestGenerateRoomRentBillFreezesAllMeterReadingData(t *testing.T) {
	itemRemark, readingRemark, reader := "水表定义", "本月读数备注", testAccountID20
	previous, waterUsage, powerUsage, zeroUsage := 7.0, 3.5, 2.0, 0.0
	readingTime := time.Date(2026, 6, 10, 8, 30, 0, 0, time.UTC)
	water := &domainmeteritem.MeterItem{
		ID: 1, AccountID: testAccountID20, Code: "水", Unit: "m³", UnitPrice: 12.345,
		CalcType: domainmeteritem.CalcTypeDifference, Remark: &itemRemark,
	}
	power := &domainmeteritem.MeterItem{
		ID: 2, AccountID: testAccountID20, Code: "电", Unit: "kWh", UnitPrice: 100.005,
		CalcType: domainmeteritem.CalcTypeDirect,
	}
	zero := &domainmeteritem.MeterItem{
		ID: 3, AccountID: testAccountID20, Code: "零用量", Unit: "次", UnitPrice: 50,
		CalcType: domainmeteritem.CalcTypeDirect,
	}
	waterReading := &domainmeterreading.MeterReading{
		ID: 21, RoomID: 11, MeterItemID: water.ID, PeriodMonth: "2026-06",
		PreviousReading: &previous, CurrentReading: 10.5, UsageAmount: &waterUsage,
		ReadingTime: readingTime, ReaderID: &reader, Source: domainmeterreading.SourceManual,
		Remark: &readingRemark,
	}
	room := &domainroom.Room{
		ID: 11, UnitID: 7, Price: 100000,
		MeterItems: []*domainmeteritem.MeterItem{water, power, zero},
	}
	repo := &fakeRentBillRepo{byID: make(map[int]*domainrentbill.RentBill)}
	service := newTestService(repo, []*domainroom.Room{room},
		[]*domainunit.Unit{{ID: 7, AccountID: testAccountID20}},
		map[string][]*domainmeterreading.MeterReading{
			meterReadingKey(11, "2026-06"): {
				waterReading,
				{ID: 22, RoomID: 11, MeterItemID: power.ID, PeriodMonth: "2026-06",
					CurrentReading: 2, UsageAmount: &powerUsage, ReadingTime: readingTime,
					Source: domainmeterreading.SourceImport},
				{ID: 23, RoomID: 11, MeterItemID: zero.ID, PeriodMonth: "2026-06",
					CurrentReading: 0, UsageAmount: &zeroUsage, ReadingTime: readingTime,
					Source: domainmeterreading.SourceAuto},
			},
		}, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
	result, err := service.GenerateRoomRentBill(ctx,
		&GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatal(err)
	}
	bill := result.Saved[0]
	if len(bill.MeterReadings) != 3 || bill.MeterFee != 243 || bill.TotalAmount != 100243 {
		t.Fatalf("generated frozen bill = %+v", bill)
	}
	line := bill.MeterReadings[0]
	if line.ID != 21 || line.RoomID != 11 || line.PeriodMonth != "2026-06" ||
		line.MeterItem.Code != "水" || line.MeterItem.Unit != "m³" ||
		line.MeterItem.UnitPrice != 12.345 || line.MeterItem.CalcType != "DIFFERENCE" ||
		line.PreviousReading == nil || *line.PreviousReading != 7 || line.CurrentReading != 10.5 ||
		line.UsageAmount == nil || *line.UsageAmount != 3.5 || !line.ReadingTime.Equal(readingTime) ||
		line.ReaderID == nil || *line.ReaderID != testAccountID20 || line.Source != "MANUAL" ||
		line.Remark == nil || *line.Remark != "本月读数备注" || line.Fee != 43 || !line.IsRecorded {
		t.Fatalf("incomplete snapshot = %+v", line)
	}
	if bill.MeterReadings[1].PreviousReading != nil || bill.MeterReadings[2].CurrentReading != 0 ||
		bill.MeterReadings[2].Fee != 0 || *bill.MeterReadings[2].UsageAmount != 0 {
		t.Fatalf("direct or zero readings lost: %+v", bill.MeterReadings)
	}
	totalFee := 0
	for _, snapshot := range bill.MeterReadings {
		totalFee += snapshot.Fee
	}
	if totalFee != bill.MeterFee {
		t.Fatalf("snapshot sum = %d, meter fee = %d", totalFee, bill.MeterFee)
	}
	expected, err := json.Marshal(ToRentBillDetailResponse(bill))
	if err != nil {
		t.Fatal(err)
	}
	repo.byID[bill.ID] = bill
	water.Code, water.Unit, water.UnitPrice, water.CalcType = "改名", "新单位", 999, "DIRECT"
	previous, waterUsage, itemRemark, readingRemark, reader = 99, 200, "改定义", "改备注", testAccountID30
	waterReading.CurrentReading, waterReading.Source, waterReading.ReadingTime = 500, "AUTO", time.Now()
	room.MeterItems = nil
	liveRepo := service.meterReadingRepo.(*fakeMeterReadingRepo)
	liveRepo.listErr = errors.New("live readings unavailable")
	stored, err := service.GetRentBill(ctx, bill.ID)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := json.Marshal(ToRentBillDetailResponse(stored))
	if err != nil || !bytes.Equal(actual, expected) {
		t.Fatalf("snapshot changed after original edits/unbinding: %s, want %s, err=%v", actual, expected, err)
	}
	if liveRepo.listCalls != 1 {
		t.Fatalf("live reads = %d, want exactly one during generation", liveRepo.listCalls)
	}
}

func TestRentBillDetailDistinguishesLegacySnapshotFromNoMeterItems(t *testing.T) {
	repo := &fakeRentBillRepo{}
	service := newTestService(repo, []*domainroom.Room{{ID: 11, UnitID: 7, Price: 100}},
		[]*domainunit.Unit{{ID: 7, AccountID: testAccountID20}}, nil, &fakeRbacService{})
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
	result, err := service.GenerateRoomRentBill(ctx,
		&GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Saved[0].MeterReadings == nil || len(result.Saved[0].MeterReadings) != 0 {
		t.Fatalf("no-item bill snapshot = %#v, want non-nil empty", result.Saved[0].MeterReadings)
	}
	for _, test := range []struct {
		name string
		bill *domainrentbill.RentBill
		want string
	}{
		{name: "legacy", bill: &domainrentbill.RentBill{ID: 99}, want: "null"},
		{name: "no items", bill: result.Saved[0], want: "[]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(ToRentBillDetailResponse(test.bill))
			if err != nil {
				t.Fatal(err)
			}
			var response map[string]json.RawMessage
			if err := json.Unmarshal(data, &response); err != nil {
				t.Fatal(err)
			}
			if string(response["meter_readings"]) != test.want {
				t.Fatalf("meter_readings = %s, want %s", response["meter_readings"], test.want)
			}
			listData, err := json.Marshal(ToRentBillResponse(test.bill))
			if err != nil || bytes.Contains(listData, []byte("meter_readings")) {
				t.Fatalf("list response changed: %s, %v", listData, err)
			}
		})
	}
}

func TestGetRentBillSnapshotChecksOnlyBillPermission(t *testing.T) {
	for _, denied := range []bool{false, true} {
		rbac := &fakeRbacService{}
		if denied {
			rbac.err = domainrbac.ErrForbidden
		}
		bill := &domainrentbill.RentBill{ID: 99, RoomID: 11,
			MeterReadings: []domainrentbill.MeterReadingSnapshot{{ID: 31, Fee: 100}}}
		service := newTestService(&fakeRentBillRepo{byID: map[int]*domainrentbill.RentBill{99: bill}},
			[]*domainroom.Room{{ID: 11, UnitID: 7}},
			[]*domainunit.Unit{{ID: 7, AccountID: testAccountID20}}, nil, rbac)
		liveRepo := service.meterReadingRepo.(*fakeMeterReadingRepo)
		liveRepo.listErr = errors.New("live readings must not be queried")
		ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID30)
		got, err := service.GetRentBill(ctx, 99)
		if denied {
			if !errors.Is(err, domainrbac.ErrForbidden) || got != nil {
				t.Fatalf("denied detail = %+v, %v", got, err)
			}
		} else if err != nil || len(got.MeterReadings) != 1 || got.MeterReadings[0].ID != 31 {
			t.Fatalf("granted detail = %+v, %v", got, err)
		}
		if rbac.permissionCode != domainrbac.PermissionRentBillView ||
			rbac.accountID != testAccountID30 || rbac.unitID != 7 || liveRepo.listCalls != 0 {
			t.Fatalf("detail permission/live access: rbac=%+v, live reads=%d", rbac, liveRepo.listCalls)
		}
	}
}

func TestRentBillDetailResponsePreservesZeroReadings(t *testing.T) {
	zero := 0.0
	bill := &domainrentbill.RentBill{MeterReadings: []domainrentbill.MeterReadingSnapshot{{
		PreviousReading: &zero, CurrentReading: 0, UsageAmount: &zero, Fee: 0, IsRecorded: true,
	}}}
	data, err := json.Marshal(ToRentBillDetailResponse(bill))
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		MeterReadings []map[string]json.RawMessage `json:"meter_readings"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"previous_reading", "current_reading", "usage_amount", "fee"} {
		if string(response.MeterReadings[0][field]) != "0" {
			t.Fatalf("%s = %s, want preserved zero", field, response.MeterReadings[0][field])
		}
	}
}

func TestGenerateRoomRentBillSkipsClosedSnapshots(t *testing.T) {
	for _, status := range []string{domainrentbill.StatusPaid, domainrentbill.StatusCancelled} {
		t.Run(status, func(t *testing.T) {
			existing := &domainrentbill.RentBill{ID: 99, RoomID: 11, PeriodMonth: "2026-06", Status: status,
				MeterReadings: []domainrentbill.MeterReadingSnapshot{{ID: 31, Fee: 100}}}
			repo := &fakeRentBillRepo{byRoomPeriod: map[string]*domainrentbill.RentBill{
				rentBillKey(11, "2026-06"): existing,
			}}
			service := newTestService(repo, []*domainroom.Room{{ID: 11, UnitID: 7}},
				[]*domainunit.Unit{{ID: 7, AccountID: testAccountID20}}, nil, &fakeRbacService{})
			liveRepo := service.meterReadingRepo.(*fakeMeterReadingRepo)
			liveRepo.listErr = errors.New("closed bill must not read new records")
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID20)
			result, err := service.GenerateRoomRentBill(ctx,
				&GenerateRoomRentBillCommand{RoomID: 11, PeriodMonth: "2026-06"})
			if err != nil || len(result.Saved) != 0 || len(result.Skipped) != 1 ||
				result.Skipped[0].MeterReadings[0].ID != 31 || liveRepo.listCalls != 0 || len(repo.saved) != 0 {
				t.Fatalf("closed bill regenerated: result=%+v, err=%v, live reads=%d", result, err, liveRepo.listCalls)
			}
		})
	}
}
