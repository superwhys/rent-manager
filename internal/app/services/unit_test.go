package services

import (
	"errors"
	"testing"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

func TestUnitServiceListUnitsDefaultsToAllScope(t *testing.T) {
	repo := &fakeUnitRepository{}
	service := NewUnitService(repo, nil)

	_, err := service.ListUnits(tokenContext(1), &requests.ListUnitsQuery{})
	if err != nil {
		t.Fatalf("expected list units, got %v", err)
	}
	if repo.listQuery.AccountID != 1 {
		t.Fatalf("expected account id 1, got %d", repo.listQuery.AccountID)
	}
	if repo.listQuery.Scope != domainunit.ListScopeAll {
		t.Fatalf("expected all scope, got %q", repo.listQuery.Scope)
	}
}

func TestUnitServiceListUnitsRejectsInvalidScope(t *testing.T) {
	service := NewUnitService(&fakeUnitRepository{}, nil)

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
	}, nil)

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
	service := NewUnitService(repo, NewRBACService(&fakeRBACRepository{}, nil))

	_, err := service.GetUnit(tokenContext(2), 10)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without unit.view, got %v", err)
	}

	service = NewUnitService(repo, NewRBACService(&fakeRBACRepository{
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
	service := NewUnitService(repo, NewRBACService(&fakeRBACRepository{
		unitPermissions: map[int]map[int]map[string]bool{
			2: {10: {domainrbac.PermissionUnitView: true}},
		},
	}, nil))

	_, err := service.UpdateUnit(tokenContext(2), cmd)
	if !errors.Is(err, domainrbac.ErrForbidden) {
		t.Fatalf("expected forbidden without unit.update, got %v", err)
	}

	service = NewUnitService(repo, NewRBACService(&fakeRBACRepository{
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
