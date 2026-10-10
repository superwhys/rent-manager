package home

import (
	"context"
	"errors"
	"testing"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestGetHomeSummaryFiltersAggregatesByPermission(t *testing.T) {
	tests := []struct {
		name        string
		canRooms    bool
		canBills    bool
		canReadings bool
	}{
		{name: "meter reading only", canReadings: true},
		{name: "room and meter reading view", canRooms: true, canReadings: true},
		{name: "rent bill and meter reading view", canBills: true, canReadings: true},
		{name: "all view permissions", canRooms: true, canBills: true, canReadings: true},
		{name: "no view permissions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			denied := make(map[string]error)
			if !tt.canRooms {
				denied[domainrbac.PermissionRoomView] = domainrbac.ErrForbidden
			}
			if !tt.canBills {
				denied[domainrbac.PermissionRentBillView] = domainrbac.ErrForbidden
			}
			if !tt.canReadings {
				denied[domainrbac.PermissionMeterReadingView] = domainrbac.ErrForbidden
			}
			service := newHomeTestService(
				[]*domainunit.Unit{{ID: 1, AccountID: testAccountID20, Name: "授权楼栋"}},
				[]*domainroom.Room{
					{ID: 11, UnitID: 1, MeterItems: []*domainmeteritem.MeterItem{{ID: 1}}},
					{ID: 12, UnitID: 1, IsActive: true},
				},
				nil,
				[]*domainrentbill.RentBill{
					{ID: 1, RoomID: 11, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 123456},
					{ID: 2, RoomID: 11, PeriodMonth: "2026-05", Status: domainrentbill.StatusPaid, TotalAmount: 100000},
					{ID: 3, RoomID: 11, PeriodMonth: "2026-06", Status: domainrentbill.StatusUnpaid, TotalAmount: 1000},
				},
				&fakeHomeRbacService{denied: denied},
			)
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID30)
			summary, err := service.GetHomeSummary(ctx)
			if err != nil {
				t.Fatal(err)
			}
			wantRooms, wantRented, wantVacant := 0, 0, 0
			if tt.canRooms {
				wantRooms, wantRented, wantVacant = 2, 1, 1
			}
			if summary.RoomCount != wantRooms || summary.RentedRoomCount != wantRented || summary.VacantRoomCount != wantVacant {
				t.Errorf("room statistics = %d/%d/%d, want %d/%d/%d", summary.RoomCount, summary.RentedRoomCount, summary.VacantRoomCount, wantRooms, wantRented, wantVacant)
			}
			wantIncome, wantPrevious, wantUnpaid := 0, 0, 0
			if tt.canBills {
				wantIncome, wantPrevious, wantUnpaid = 123456, 100000, 1
			}
			if summary.CurrentMonthIncomeCents != wantIncome || summary.PreviousMonthIncomeCents != wantPrevious || summary.CurrentMonthUnpaidBillCount != wantUnpaid || summary.MonthlyIncomeGrowthCents != wantIncome-wantPrevious {
				t.Errorf("financial statistics = %+v, want income=%d, previous=%d, unpaid=%d", summary, wantIncome, wantPrevious, wantUnpaid)
			}
			var wantTodos []Todo
			if tt.canReadings {
				wantTodos = []Todo{{Title: "授权楼栋", Description: "还有 1 间房未抄表", Tag: todoTagMeterReading}}
			}
			assertHomeTodos(t, summary.Todos, wantTodos)
		})
	}
}

func TestGetHomeSummaryDoesNotCombineUnauthorizedUnitData(t *testing.T) {
	service := newHomeTestService(
		[]*domainunit.Unit{
			{ID: 1, AccountID: testAccountID20},
			{ID: 2, AccountID: testAccountID20},
		},
		[]*domainroom.Room{
			{ID: 11, UnitID: 1, IsActive: true},
			{ID: 12, UnitID: 1, IsActive: true},
			{ID: 21, UnitID: 2, IsActive: true},
		},
		nil,
		[]*domainrentbill.RentBill{
			{ID: 1, RoomID: 11, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 999999},
			{ID: 2, RoomID: 21, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid, TotalAmount: 25000},
		},
		&homeUnitPermissions{allowed: map[int]string{1: domainrbac.PermissionRoomView, 2: domainrbac.PermissionRentBillView}},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID30)
	summary, err := service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.RoomCount != 2 || summary.VacantRoomCount != 2 || summary.CurrentMonthIncomeCents != 25000 {
		t.Fatalf("aggregates include unauthorized unit data: %+v", summary)
	}
}

func TestGetHomeSummaryReturnsPermissionLookupErrors(t *testing.T) {
	wantErr := errors.New("permission lookup failed")
	service := newHomeTestService(
		[]*domainunit.Unit{{ID: 1, AccountID: testAccountID20}},
		nil, nil, nil,
		&fakeHomeRbacService{denied: map[string]error{domainrbac.PermissionRoomView: wantErr}},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID30)
	summary, err := service.GetHomeSummary(ctx)
	if !errors.Is(err, wantErr) || summary != nil {
		t.Fatalf("GetHomeSummary() = %+v, %v; want nil, permission lookup error", summary, err)
	}
}

type homeUnitPermissions struct {
	allowed map[int]string
}

func (s *homeUnitPermissions) RequireUnitPermission(_ context.Context, _ string, unitID int, permission string) error {
	if s.allowed[unitID] == permission {
		return nil
	}
	return domainrbac.ErrForbidden
}
