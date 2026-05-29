package services

import (
	"errors"
	"testing"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

func TestUnitServiceListUnitsDefaultsToAllScope(t *testing.T) {
	repo := &fakeUnitRepository{
		listUnits: []*domainunit.Unit{
			{ID: 10, AccountID: 1, Name: "A"},
			{ID: 20, AccountID: 1, Name: "B"},
		},
	}
	roomRepo := &fakeRoomRepository{
		rooms: map[int]*domainroom.Room{
			1: {ID: 1, UnitID: 10, IsActive: true},
			2: {ID: 2, UnitID: 10, IsActive: false},
			3: {ID: 3, UnitID: 20, IsActive: true},
			4: {ID: 4, UnitID: 20, IsActive: true},
		},
	}
	service := NewUnitService(repo, roomRepo, nil)

	summaries, err := service.ListUnits(tokenContext(1), &requests.ListUnitsQuery{})
	if err != nil {
		t.Fatalf("expected list units, got %v", err)
	}
	if repo.listQuery.AccountID != 1 {
		t.Fatalf("expected account id 1, got %d", repo.listQuery.AccountID)
	}
	if repo.listQuery.Scope != domainunit.ListScopeAll {
		t.Fatalf("expected all scope, got %q", repo.listQuery.Scope)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].RoomCount != 2 || summaries[0].RentedCount != 1 {
		t.Fatalf("unexpected room summary: %#v", summaries[0])
	}
	if summaries[0].OccupancyStatus != "vacancy" {
		t.Fatalf("expected vacancy status, got %q", summaries[0].OccupancyStatus)
	}
	if summaries[1].RoomCount != 2 || summaries[1].RentedCount != 2 {
		t.Fatalf("unexpected full room summary: %#v", summaries[1])
	}
	if summaries[1].OccupancyStatus != "full" {
		t.Fatalf("expected full status, got %q", summaries[1].OccupancyStatus)
	}
	if summaries[0].MonthlyRentAmount == "" || summaries[1].MonthlyRentAmount == "" {
		t.Fatalf("expected mock monthly rent amount")
	}
}

func TestUnitServiceListUnitsRejectsInvalidScope(t *testing.T) {
	service := NewUnitService(&fakeUnitRepository{}, &fakeRoomRepository{}, nil)

	_, err := service.ListUnits(tokenContext(1), &requests.ListUnitsQuery{Scope: "bad"})
	if !errors.Is(err, domainunit.ErrInvalidListScope) {
		t.Fatalf("expected invalid list scope, got %v", err)
	}
}

func TestUnitServiceGetUnitAllowsOwnerWithoutRBAC(t *testing.T) {
	service := NewUnitService(&fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}, nil, nil)

	unit, err := service.GetUnit(tokenContext(1), 10)
	if err != nil {
		t.Fatalf("owner should get own unit without rbac, got %v", err)
	}
	if unit.AccessType != domainunit.AccessTypeOwned {
		t.Fatalf("expected owned access type, got %q", unit.AccessType)
	}
}

func TestUnitServiceGetUnitRequiresViewPermissionForGrantedUnit(t *testing.T) {
	repo := &fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}
	service := NewUnitService(repo, nil, NewRBACService(&fakeRBACRepository{}, nil))

	_, err := service.GetUnit(tokenContext(2), 10)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without unit.view, got %v", err)
	}

	service = NewUnitService(repo, nil, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionUnitView: true}},
		},
	}, nil))
	unit, err := service.GetUnit(tokenContext(2), 10)
	if err != nil {
		t.Fatalf("expected granted unit.view, got %v", err)
	}
	if unit.AccessType != domainunit.AccessTypeGranted {
		t.Fatalf("expected granted access type, got %q", unit.AccessType)
	}
}

func TestUnitServiceUpdateUnitRequiresUpdatePermissionForGrantedUnit(t *testing.T) {
	repo := &fakeUnitRepository{
		units: map[int]*domainunit.Unit{
			10: {ID: 10, AccountID: 1, Name: "A", Address: "addr", City: "city", Province: "province"},
		},
	}
	cmd := &requests.UpdateUnitCommand{
		UnitID:   10,
		Name:     "B",
		Address:  "addr2",
		City:     "city2",
		Province: "province2",
	}
	service := NewUnitService(repo, nil, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionUnitView: true}},
		},
	}, nil))

	_, err := service.UpdateUnit(tokenContext(2), cmd)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without unit.update, got %v", err)
	}

	service = NewUnitService(repo, nil, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionUnitUpdate: true}},
		},
	}, nil))
	unit, err := service.UpdateUnit(tokenContext(2), cmd)
	if err != nil {
		t.Fatalf("expected granted unit.update, got %v", err)
	}
	if unit.AccountID != 1 {
		t.Fatalf("expected owner account id preserved, got %d", unit.AccountID)
	}
	if unit.AccessType != domainunit.AccessTypeGranted {
		t.Fatalf("expected granted access type, got %q", unit.AccessType)
	}
}
