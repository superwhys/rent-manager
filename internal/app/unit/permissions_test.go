package unit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	appcore "github.com/superwhys/rent-manager/internal/app"
	domainrbac "github.com/superwhys/rent-manager/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
)

func TestListUnitsFiltersFieldsByPermission(t *testing.T) {
	tests := []struct {
		name        string
		owned       bool
		permissions []string
		details     bool
		rooms       bool
		bills       bool
	}{
		{name: "owner", owned: true, details: true, rooms: true, bills: true},
		{name: "all view permissions", permissions: []string{domainrbac.PermissionUnitView, domainrbac.PermissionRoomView, domainrbac.PermissionRentBillView}, details: true, rooms: true, bills: true},
		{name: "unit view only", permissions: []string{domainrbac.PermissionUnitView}, details: true},
		{name: "unit and room view", permissions: []string{domainrbac.PermissionUnitView, domainrbac.PermissionRoomView}, details: true, rooms: true},
		{name: "unit and bill view", permissions: []string{domainrbac.PermissionUnitView, domainrbac.PermissionRentBillView}, details: true, bills: true},
		{name: "meter reading entry", permissions: []string{domainrbac.PermissionMeterReadingView}},
		{name: "revoked unit view", permissions: []string{domainrbac.PermissionRoomView, domainrbac.PermissionRentBillView}},
		{name: "empty role"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerID := "0000000000000000000000000000000000000000000000000000000000000014"
			accessType := domainunit.AccessTypeGranted
			if tt.owned {
				ownerID, accessType = testAccountID10, domainunit.AccessTypeOwned
			}
			unit := &domainunit.Unit{
				ID: 1, AccountID: ownerID, Name: "楼栋入口", AccessType: accessType,
				Address: "保密地址", City: "城市", Province: "省份",
				DefaultMeterItems: []*domainunit.MeterItem{{ID: 2, Code: "water", UnitPrice: 5}},
				Config:            domainunit.UnitConfig{MeterReadingDay: 20},
			}
			allowed := make(map[string]bool)
			for _, permission := range tt.permissions {
				allowed[permission] = true
			}
			roomRepo := &unitListRoomRepo{}
			billRepo := &unitListRentBillRepo{}
			service := &Service{
				repo: &unitListRepo{units: []*domainunit.Unit{unit}}, roomRepo: roomRepo,
				rentBillRepo: billRepo, now: time.Now,
				rbacService: &unitListRbacService{allowed: allowed},
			}
			ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
			summaries, err := service.ListUnits(ctx, &ListUnitsQuery{})
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(ToUnitListResponse(summaries))
			if err != nil {
				t.Fatal(err)
			}
			var responses []map[string]any
			if err := json.Unmarshal(data, &responses); err != nil {
				t.Fatal(err)
			}
			if len(responses) != 1 {
				t.Fatalf("responses = %s, want one discoverable unit", data)
			}
			response := responses[0]
			if response["id"] != float64(1) || response["name"] != "楼栋入口" || response["access_type"] != accessType || response["is_owner"] != tt.owned {
				t.Errorf("unit identity = %s", data)
			}
			for _, key := range []string{"account_id", "address", "city", "province", "default_meter_items", "json_extend"} {
				if _, present := response[key]; present != tt.details {
					t.Errorf("field %s present=%t, want %t: %s", key, present, tt.details, data)
				}
			}
			for _, key := range []string{"room_count", "rented_count", "occupancy_status"} {
				if _, present := response[key]; present != tt.rooms {
					t.Errorf("field %s present=%t, want %t: %s", key, present, tt.rooms, data)
				}
			}
			for _, key := range []string{"monthly_rent_amount", "current_month_received_amount"} {
				if _, present := response[key]; present != tt.bills {
					t.Errorf("field %s present=%t, want %t: %s", key, present, tt.bills, data)
				}
			}
			if tt.bills && response["current_month_received_amount"] != float64(0) {
				t.Errorf("current_month_received_amount = %v, want 0", response["current_month_received_amount"])
			}
			if !tt.rooms && roomRepo.countCalls != 0 {
				t.Errorf("room count queried %d times without list statistics permission", roomRepo.countCalls)
			}
			wantBillCalls := 0
			if tt.bills {
				wantBillCalls = 1
			}
			if len(billRepo.calls) != wantBillCalls {
				t.Errorf("rent bills queried %d times, want %d", len(billRepo.calls), wantBillCalls)
			}
		})
	}
}

func TestListUnitsReturnsPermissionLookupErrors(t *testing.T) {
	wantErr := errors.New("permission lookup failed")
	service := &Service{
		repo:        &unitListRepo{units: []*domainunit.Unit{{ID: 1, AccountID: "other"}}},
		rbacService: &unitListRbacService{err: wantErr},
	}
	ctx := appcore.SetAccountIDToContext(context.Background(), testAccountID10)
	summaries, err := service.ListUnits(ctx, &ListUnitsQuery{})
	if !errors.Is(err, wantErr) || summaries != nil {
		t.Fatalf("ListUnits() = %+v, %v; want nil, permission lookup error", summaries, err)
	}
}

type unitListRepo struct {
	domainunit.Repository
	units []*domainunit.Unit
}

func (r *unitListRepo) List(context.Context, domainunit.ListQuery) ([]*domainunit.Unit, error) {
	return r.units, nil
}

type unitListRoomRepo struct {
	domainroom.Repository
	countCalls int
}

func (r *unitListRoomRepo) Count(context.Context, string, domainroom.ListQuery) (int, error) {
	r.countCalls++
	return 0, nil
}

type unitListRbacService struct {
	allowed map[string]bool
	err     error
}

func (s *unitListRbacService) RequireUnitPermission(_ context.Context, _ string, _ int, permission string) error {
	if s.err != nil {
		return s.err
	}
	if s.allowed[permission] {
		return nil
	}
	return domainrbac.ErrForbidden
}
