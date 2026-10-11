package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestListUnitsCurrentMonthReceivedAmount(t *testing.T) {
	otherOwner := "0000000000000000000000000000000000000000000000000000000000000014"
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	laterPayment := now.AddDate(0, 1, 0)
	units := []*domainunit.Unit{
		{ID: 1, AccountID: testAccountID10},
		{ID: 2, AccountID: testAccountID10},
		{ID: 3, AccountID: otherOwner},
		{ID: 4, AccountID: testAccountID10},
	}
	billRepo := &unitListRentBillRepo{records: []unitListRentBillRecord{
		{accountID: testAccountID10, unitID: 1, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 12001, PaidAt: &laterPayment}},
		{accountID: testAccountID10, unitID: 1, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 99}},
		{accountID: testAccountID10, unitID: 1, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusUnpaid, TotalAmount: 99999}},
		{accountID: testAccountID10, unitID: 1, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusCancelled, TotalAmount: 88888}},
		{accountID: testAccountID10, unitID: 1, bill: &domainrentbill.RentBill{PeriodMonth: "2026-05", Status: domainrentbill.StatusPaid, TotalAmount: 77777, PaidAt: &now}},
		{accountID: testAccountID10, unitID: 2, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 23456}},
		{accountID: otherOwner, unitID: 3, bill: &domainrentbill.RentBill{PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 34567}},
	}}
	clockCalls := 0
	service := &Service{
		repo:         &unitListRepo{units: units},
		roomRepo:     &unitListRoomRepo{},
		rentBillRepo: billRepo,
		rbacService: &unitListRbacService{allowed: map[string]bool{
			domainrbac.PermissionUnitView:     true,
			domainrbac.PermissionRentBillView: true,
		}},
		now: func() time.Time {
			clockCalls++
			return now
		},
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	summaries, err := service.ListUnits(ctx, &ListUnitsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	wantAmounts := []int{12100, 23456, 34567, 0}
	if len(summaries) != len(wantAmounts) || len(billRepo.calls) != len(wantAmounts) {
		t.Fatalf("summaries=%d, bill calls=%d; want %d each", len(summaries), len(billRepo.calls), len(wantAmounts))
	}
	for i, summary := range summaries {
		if summary.CurrentMonthReceivedAmount != wantAmounts[i] {
			t.Errorf("unit %d received=%d, want %d", summary.Unit.ID, summary.CurrentMonthReceivedAmount, wantAmounts[i])
		}
		response := ToUnitSummaryResponse(summary)
		if response.CurrentMonthReceivedAmount == nil || *response.CurrentMonthReceivedAmount != wantAmounts[i] {
			t.Errorf("unit %d response received=%v, want %d", summary.Unit.ID, response.CurrentMonthReceivedAmount, wantAmounts[i])
		}
		call := billRepo.calls[i]
		wantQuery := domainrentbill.ListQuery{UnitID: units[i].ID, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid}
		if call.accountID != units[i].AccountID || call.query != wantQuery || call.ctx != ctx {
			t.Errorf("unit %d bill query=%+v, want owner=%s, query=%+v, original context", units[i].ID, call, units[i].AccountID, wantQuery)
		}
	}
	if clockCalls != 1 {
		t.Errorf("clock queried %d times, want once for a consistent list month", clockCalls)
	}
}

func TestListUnitsCurrentMonthUsesShanghaiTime(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{name: "before month boundary", now: time.Date(2026, 5, 31, 15, 59, 59, 0, time.UTC), want: "2026-05"},
		{name: "at month boundary", now: time.Date(2026, 5, 31, 16, 0, 0, 0, time.UTC), want: "2026-06"},
		{name: "before year boundary", now: time.Date(2026, 12, 31, 15, 59, 59, 0, time.UTC), want: "2026-12"},
		{name: "at year boundary", now: time.Date(2026, 12, 31, 16, 0, 0, 0, time.UTC), want: "2027-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billRepo := &unitListRentBillRepo{}
			service := &Service{
				repo:         &unitListRepo{units: []*domainunit.Unit{{ID: 1, AccountID: testAccountID10}}},
				roomRepo:     &unitListRoomRepo{},
				rentBillRepo: billRepo,
				now:          func() time.Time { return tt.now },
			}
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
			if _, err := service.ListUnits(ctx, &ListUnitsQuery{}); err != nil {
				t.Fatal(err)
			}
			if len(billRepo.calls) != 1 || billRepo.calls[0].query.PeriodMonth != tt.want {
				t.Fatalf("bill queries=%+v, want month %s", billRepo.calls, tt.want)
			}
		})
	}
}

func TestListUnitsReturnsRentBillLookupError(t *testing.T) {
	wantErr := errors.New("账单查询失败")
	service := &Service{
		repo:         &unitListRepo{units: []*domainunit.Unit{{ID: 1, AccountID: testAccountID10}}},
		roomRepo:     &unitListRoomRepo{},
		rentBillRepo: &unitListRentBillRepo{err: wantErr},
		now:          time.Now,
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	summaries, err := service.ListUnits(ctx, &ListUnitsQuery{})
	if err != wantErr || summaries != nil {
		t.Fatalf("ListUnits()=%+v, %v; want nil, original error %v", summaries, err, wantErr)
	}
}

func TestListUnitsEmptyListDoesNotQueryStatistics(t *testing.T) {
	service := &Service{repo: &unitListRepo{}}
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	summaries, err := service.ListUnits(ctx, &ListUnitsQuery{})
	if err != nil || len(summaries) != 0 {
		t.Fatalf("ListUnits()=%+v, %v; want empty list", summaries, err)
	}
}

type unitListRentBillRecord struct {
	accountID string
	unitID    int
	bill      *domainrentbill.RentBill
}

type unitListRentBillCall struct {
	ctx       context.Context
	accountID string
	query     domainrentbill.ListQuery
}

type unitListRentBillRepo struct {
	domainrentbill.Repository
	records []unitListRentBillRecord
	calls   []unitListRentBillCall
	err     error
}

func (r *unitListRentBillRepo) List(ctx context.Context, accountID string, query domainrentbill.ListQuery) ([]*domainrentbill.RentBill, error) {
	r.calls = append(r.calls, unitListRentBillCall{ctx: ctx, accountID: accountID, query: query})
	if r.err != nil {
		return nil, r.err
	}
	var bills []*domainrentbill.RentBill
	for _, record := range r.records {
		if record.accountID == accountID && record.unitID == query.UnitID && record.bill.PeriodMonth == query.PeriodMonth && record.bill.Status == query.Status {
			bills = append(bills, record.bill)
		}
	}
	return bills, nil
}
