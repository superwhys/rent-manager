package home

import (
	"context"
	"strings"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestHomeReadingTodosUseCurrentMonthBeforeOnAndAfterReadingDay(t *testing.T) {
	for _, tt := range []struct {
		name string
		day  int
		want []string
	}{
		{name: "before reading day", day: 19, want: []string{"2026-10", "计划", "2026-10-20"}},
		{name: "on reading day", day: 20, want: []string{"2026-10", "今日"}},
		{name: "after reading day", day: 21, want: []string{"2026-10", "逾期"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := newHomeTestServiceAt(
				time.Date(2026, time.October, tt.day, 9, 0, 0, 0, time.UTC),
				[]*domainunit.Unit{{
					ID: 1, AccountID: testAccountID10, Name: "当月抄表楼栋",
					Config: domainunit.UnitConfig{MeterReadingDay: 20},
				}},
				[]*domainroom.Room{{ID: 11, UnitID: 1, MeterItemIDs: []int{1}}},
				[]*domainmeterreading.MeterReading{
					{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-09", UsageAmount: new(1.0)},
				}, nil,
				&fakeHomeRbacService{denied: map[string]error{
					domainrbac.PermissionRentBillView: domainrbac.ErrForbidden,
				}},
			)
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

			summary, err := service.GetHomeSummary(ctx)
			if err != nil {
				t.Fatal(err)
			}
			assertReadingDayTodoContains(t, summary.Todos, todoTagMeterReading, tt.want...)

			service.meterReadingRepo.(*fakeHomeMeterReadingRepo).readings = append(
				service.meterReadingRepo.(*fakeHomeMeterReadingRepo).readings,
				&domainmeterreading.MeterReading{
					RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-10", UsageAmount: new(0.0),
				},
			)
			summary, err = service.GetHomeSummary(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(summary.Todos) != 0 {
				t.Fatalf("completed current-month reading has todos: %#v", summary.Todos)
			}
		})
	}
}

func TestHomeReadingTodosRespectCalendarAndBeijingTime(t *testing.T) {
	for _, tt := range []struct {
		name       string
		now        time.Time
		readingDay int
		month      string
		want       []string
	}{
		{
			name: "short month before final day",
			now:  time.Date(2026, time.April, 29, 9, 0, 0, 0, time.UTC), readingDay: 31,
			month: "2026-04", want: []string{"计划", "2026-04-30"},
		},
		{
			name: "short month on final day",
			now:  time.Date(2026, time.April, 30, 9, 0, 0, 0, time.UTC), readingDay: 31,
			month: "2026-04", want: []string{"今日"},
		},
		{
			name: "December still uses December",
			now:  time.Date(2026, time.December, 31, 9, 0, 0, 0, time.UTC), readingDay: 20,
			month: "2026-12", want: []string{"逾期"},
		},
		{
			name: "new year uses January",
			now:  time.Date(2026, time.December, 31, 16, 0, 0, 0, time.UTC), readingDay: 20,
			month: "2027-01", want: []string{"计划", "2027-01-20"},
		},
		{
			name: "UTC date before Beijing reading day",
			now:  time.Date(2026, time.October, 19, 15, 59, 59, 0, time.UTC), readingDay: 20,
			month: "2026-10", want: []string{"计划", "2026-10-20"},
		},
		{
			name: "UTC date becomes Beijing reading day",
			now:  time.Date(2026, time.October, 19, 16, 0, 0, 0, time.UTC), readingDay: 20,
			month: "2026-10", want: []string{"今日"},
		},
		{
			name: "UTC month becomes Beijing October",
			now:  time.Date(2026, time.September, 30, 16, 0, 0, 0, time.UTC), readingDay: 20,
			month: "2026-10", want: []string{"计划", "2026-10-20"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := newHomeTestServiceAt(
				tt.now,
				[]*domainunit.Unit{{
					ID: 1, AccountID: testAccountID10, Name: "日历边界楼栋",
					Config: domainunit.UnitConfig{MeterReadingDay: tt.readingDay},
				}},
				[]*domainroom.Room{{ID: 11, UnitID: 1, MeterItemIDs: []int{1}}}, nil, nil,
				&fakeHomeRbacService{denied: map[string]error{
					domainrbac.PermissionRentBillView: domainrbac.ErrForbidden,
				}},
			)
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

			summary, err := service.GetHomeSummary(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if summary.CurrentMonth != tt.month {
				t.Fatalf("current month = %q, want %q", summary.CurrentMonth, tt.month)
			}
			assertReadingDayTodoContains(t, summary.Todos, todoTagMeterReading, append([]string{tt.month}, tt.want...)...)
		})
	}
}

func TestHomeReadingTodosKeepHistoricalPartialReadingsUntilComplete(t *testing.T) {
	service := newHomeTestServiceAt(
		time.Date(2026, time.October, 19, 9, 0, 0, 0, time.UTC),
		[]*domainunit.Unit{{
			ID: 1, AccountID: testAccountID10, Name: "历史抄表楼栋",
			Config: domainunit.UnitConfig{MeterReadingDay: 20},
		}},
		[]*domainroom.Room{{ID: 11, UnitID: 1, MeterItemIDs: []int{1, 2}}},
		[]*domainmeterreading.MeterReading{
			{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-09", UsageAmount: new(1.0)},
			{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
			{RoomID: 11, MeterItemID: 2, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
		}, nil,
		&fakeHomeRbacService{denied: map[string]error{
			domainrbac.PermissionRentBillView: domainrbac.ErrForbidden,
		}},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	summary, err := service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertReadingDayTodoContains(t, summary.Todos, todoTagMeterReading, "2026-09", "1 间房", "逾期")

	service.meterReadingRepo.(*fakeHomeMeterReadingRepo).readings = append(
		service.meterReadingRepo.(*fakeHomeMeterReadingRepo).readings,
		&domainmeterreading.MeterReading{
			RoomID: 11, MeterItemID: 2, PeriodMonth: "2026-09", UsageAmount: new(0.0),
		},
	)
	summary, err = service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Todos) != 0 {
		t.Fatalf("completed historical reading has todos: %#v", summary.Todos)
	}
}

func TestHomeReadingTodosDoNotInferHistoricalMissingItemsAfterBilling(t *testing.T) {
	for _, status := range []string{
		domainrentbill.StatusUnpaid,
		domainrentbill.StatusPaid,
		domainrentbill.StatusCancelled,
	} {
		t.Run(status, func(t *testing.T) {
			service := newHomeTestServiceAt(
				time.Date(2026, time.October, 19, 9, 0, 0, 0, time.UTC),
				[]*domainunit.Unit{{
					ID: 1, AccountID: testAccountID10, Name: "已出账楼栋",
					Config: domainunit.UnitConfig{MeterReadingDay: 20},
				}},
				[]*domainroom.Room{{ID: 11, UnitID: 1, MeterItemIDs: []int{1, 2}}},
				[]*domainmeterreading.MeterReading{
					{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-09", UsageAmount: new(1.0)},
					{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
					{RoomID: 11, MeterItemID: 2, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
				},
				[]*domainrentbill.RentBill{{ID: 1, RoomID: 11, PeriodMonth: "2026-09", Status: status}},
				&fakeHomeRbacService{denied: map[string]error{
					domainrbac.PermissionRentBillView: domainrbac.ErrForbidden,
				}},
			)
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

			summary, err := service.GetHomeSummary(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(summary.Todos) != 0 {
				t.Fatalf("historical billed month has inferred reading todos: %#v", summary.Todos)
			}
		})
	}
}

func TestHomeReadingTodosPrioritizeCurrentMonthWithinHomeLimit(t *testing.T) {
	now := time.Date(2026, time.October, 19, 9, 0, 0, 0, time.UTC)
	readings := []*domainmeterreading.MeterReading{
		{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
		{RoomID: 11, MeterItemID: 2, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
	}
	for offset := 1; offset <= 10; offset++ {
		readings = append(readings, &domainmeterreading.MeterReading{
			RoomID: 11, MeterItemID: 1,
			PeriodMonth: now.AddDate(0, -offset, 0).Format("2006-01"), UsageAmount: new(1.0),
		})
	}
	service := newHomeTestServiceAt(
		now,
		[]*domainunit.Unit{
			{
				ID: 1, AccountID: testAccountID10, Name: "历史抄表楼栋",
				Config: domainunit.UnitConfig{MeterReadingDay: 20},
			},
			{
				ID: 2, AccountID: testAccountID10, Name: "当月抄表楼栋",
				Config: domainunit.UnitConfig{MeterReadingDay: 20},
			},
		},
		[]*domainroom.Room{
			{ID: 11, UnitID: 1, MeterItemIDs: []int{1, 2}},
			{ID: 21, UnitID: 2, MeterItemIDs: []int{1}},
		},
		readings, nil,
		&fakeHomeRbacService{denied: map[string]error{
			domainrbac.PermissionRentBillView: domainrbac.ErrForbidden,
		}},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	summary, err := service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Todos) != 10 {
		t.Fatalf("todo count = %d, want 10; todos = %#v", len(summary.Todos), summary.Todos)
	}
	if summary.Todos[0].Title != "当月抄表楼栋" {
		t.Fatalf("first todo = %#v, want second unit's current-month reading", summary.Todos[0])
	}
	assertReadingDayTodoContains(t, summary.Todos[:1], todoTagMeterReading, "2026-10", "计划", "2026-10-20")
}

func TestHomeManualBillingTodosUseCompletedReadingsWithoutWaitingForReadingDay(t *testing.T) {
	for _, day := range []int{19, 20, 21} {
		service := newHomeTestServiceAt(
			time.Date(2026, time.October, day, 9, 0, 0, 0, time.UTC),
			[]*domainunit.Unit{{
				ID: 1, AccountID: testAccountID10, Name: "手工出账楼栋",
				Config: domainunit.UnitConfig{MeterReadingDay: 20},
			}},
			[]*domainroom.Room{{ID: 11, UnitID: 1, MeterItemIDs: []int{1}}},
			[]*domainmeterreading.MeterReading{
				{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-09", UsageAmount: new(1.0)},
				{RoomID: 11, MeterItemID: 1, PeriodMonth: "2026-10", UsageAmount: new(1.0)},
			}, nil, &fakeHomeRbacService{},
		)
		ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
		summary, err := service.GetHomeSummary(ctx)
		if err != nil {
			t.Fatal(err)
		}
		billTodos := readingDayTodosByTag(summary.Todos, todoTagBillGeneration)
		if len(billTodos) != 2 {
			t.Fatalf("day %d: bill todo count = %d, want 2; todos = %#v", day, len(billTodos), summary.Todos)
		}
		for i, month := range []string{"2026-09", "2026-10"} {
			assertReadingDayTodoContains(t, billTodos[i:i+1], todoTagBillGeneration, month, "1 间房", "待生成账单")
		}
	}
}

func TestHomeRentCollectionTodosIncludeAllUnpaidMonthsBeforeReadingDay(t *testing.T) {
	service := newHomeTestServiceAt(
		time.Date(2026, time.October, 19, 9, 0, 0, 0, time.UTC),
		[]*domainunit.Unit{{
			ID: 1, AccountID: testAccountID10, Name: "跨月收租楼栋",
			Config: domainunit.UnitConfig{MeterReadingDay: 20},
		}},
		[]*domainroom.Room{{ID: 11, UnitID: 1}, {ID: 12, UnitID: 1, IsActive: true}}, nil,
		[]*domainrentbill.RentBill{
			{ID: 1, RoomID: 11, PeriodMonth: "2026-08", Status: domainrentbill.StatusUnpaid},
			{ID: 2, RoomID: 11, PeriodMonth: "2026-09", Status: domainrentbill.StatusUnpaid},
			{ID: 3, RoomID: 11, PeriodMonth: "2026-10", Status: domainrentbill.StatusUnpaid},
			{ID: 4, RoomID: 12, PeriodMonth: "2026-07", Status: domainrentbill.StatusUnpaid},
			{ID: 5, RoomID: 11, PeriodMonth: "2026-06", Status: domainrentbill.StatusPaid},
			{ID: 6, RoomID: 11, PeriodMonth: "2026-05", Status: domainrentbill.StatusCancelled},
		},
		&fakeHomeRbacService{},
	)
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)

	summary, err := service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	collectionTodos := readingDayTodosByTag(summary.Todos, todoTagRentCollection)
	if len(collectionTodos) != 4 {
		t.Fatalf("collection todo count = %d, want 4; todos = %#v", len(collectionTodos), summary.Todos)
	}
	for _, month := range []string{"2026-07", "2026-08", "2026-09", "2026-10"} {
		var matching []Todo
		for _, todo := range collectionTodos {
			if strings.Contains(todo.Description, month) {
				matching = append(matching, todo)
			}
		}
		assertReadingDayTodoContains(t, matching, todoTagRentCollection, month, "1 笔", "租金待收")
	}

	for _, bill := range service.rentBillRepo.(*fakeHomeRentBillRepo).bills {
		if bill.Status == domainrentbill.StatusUnpaid {
			bill.Status = domainrentbill.StatusPaid
		}
	}
	summary, err = service.GetHomeSummary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(readingDayTodosByTag(summary.Todos, todoTagRentCollection)) != 0 {
		t.Fatalf("paid/cancelled bills have collection todos: %#v", summary.Todos)
	}
}

func assertReadingDayTodoContains(t *testing.T, todos []Todo, tag string, parts ...string) {
	t.Helper()
	matching := readingDayTodosByTag(todos, tag)
	if len(matching) != 1 {
		t.Fatalf("%s todo count = %d, want 1; todos = %#v", tag, len(matching), todos)
	}
	for _, part := range parts {
		if !strings.Contains(matching[0].Description, part) {
			t.Fatalf("%s todo description = %q, want to contain %q", tag, matching[0].Description, part)
		}
	}
}

func readingDayTodosByTag(todos []Todo, tag string) []Todo {
	var matching []Todo
	for _, todo := range todos {
		if todo.Tag == tag {
			matching = append(matching, todo)
		}
	}
	return matching
}
