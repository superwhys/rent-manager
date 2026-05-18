package services

import (
	"context"

	domainrbac "github.com/superwhys/rent-manager-server/internal/domain/rbac"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	"github.com/superwhys/rent-manager-server/internal/models/requests"
)

type UnitService struct {
	repo domainunit.Repository
	rbac *RBACService
}

func NewUnitService(repo domainunit.Repository, rbac *RBACService) *UnitService {
	return &UnitService{repo: repo, rbac: rbac}
}

func (s *UnitService) CreateUnit(ctx context.Context, cmd *requests.CreateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := domainunit.NewUnit(domainunit.CreateParams{
		AccountID:           accountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, unit)
}

func (s *UnitService) GetUnit(ctx context.Context, id int) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unit, err := s.requireUnitPermission(ctx, accountID, id, domainrbac.PermissionUnitView)
	if err != nil {
		return nil, err
	}

	return unit, nil
}

func (s *UnitService) ListUnits(ctx context.Context, query *requests.ListUnitsQuery) ([]*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	scope := query.Scope
	if scope == "" {
		scope = domainunit.ListScopeAll
	}

	switch scope {
	case domainunit.ListScopeOwned, domainunit.ListScopeGranted, domainunit.ListScopeAll:
	default:
		return nil, domainunit.ErrInvalidListScope
	}

	return s.repo.List(ctx, domainunit.ListQuery{
		AccountID: accountID,
		Scope:     scope,
		Name:      query.Name,
		City:      query.City,
		Province:  query.Province,
		Limit:     query.Limit,
		Offset:    query.Offset,
	})
}

func (s *UnitService) UpdateUnit(ctx context.Context, cmd *requests.UpdateUnitCommand) (*domainunit.Unit, error) {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := s.requireUnitPermission(ctx, accountID, cmd.UnitID, domainrbac.PermissionUnitUpdate)
	if err != nil {
		return nil, err
	}

	unit := &domainunit.Unit{}
	if err := unit.Update(domainunit.UpdateParams{
		ID:                  cmd.UnitID,
		AccountID:           existing.AccountID,
		Name:                cmd.Name,
		Address:             cmd.Address,
		City:                cmd.City,
		Province:            cmd.Province,
		DefaultMeterItemIDs: cmd.DefaultMeterItemIDs,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, unit)
	if err != nil {
		return nil, err
	}

	setUnitAccessType(updated, accountID)
	return updated, nil
}

func (s *UnitService) DeleteUnit(ctx context.Context, id int) error {
	accountID, err := AccountIDFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountID, id)
}

func (s *UnitService) requireUnitPermission(ctx context.Context, accountID int, unitID int, permissionCode string) (*domainunit.Unit, error) {
	unit, err := s.repo.GetByID(ctx, unitID)
	if err != nil {
		return nil, err
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
		return unit, nil
	}

	if err := s.rbac.RequireUnitPermission(ctx, accountID, unitID, permissionCode); err != nil {
		return nil, err
	}

	unit.AccessType = domainunit.AccessTypeGranted
	return unit, nil
}

func setUnitAccessType(unit *domainunit.Unit, accountID int) {
	if unit == nil {
		return
	}

	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}
}
