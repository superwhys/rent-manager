package mysql

import (
	"context"
	"time"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/predicate"

	domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"
	entclient "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent"
	entmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/meteritem"
	entrole "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/role"
	entroom "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/room"
	entunit "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/unit"
	entunitaccountrole "git.superwhys.top/rent-manager/rent-manager-server/internal/infra/persistence/ent/unitaccountrole"
)

type UnitRepository struct {
	client *entclient.Client
}

func NewUnitRepository(client *entclient.Client) *UnitRepository {
	return &UnitRepository{client: client}
}

func (r *UnitRepository) Create(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	if err := r.ensureMeterItemsInAccount(ctx, unit.AccountID, unit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}
	record, err := r.client.Unit.Create().
		SetAccountID(unit.AccountID).
		SetName(unit.Name).
		SetAddress(unit.Address).
		SetCity(unit.City).
		SetProvince(unit.Province).
		AddMeterItemIDs(unit.DefaultMeterItemIDs...).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainunit.ErrNameAlreadyExists
		}
		return nil, err
	}
	return r.Get(ctx, unit.AccountID, record.ID)
}

func (r *UnitRepository) Get(ctx context.Context, accountID int, id int) (*domainunit.Unit, error) {
	record, err := r.client.Unit.Query().
		Where(entunit.ID(id), entunit.AccountID(accountID), entunit.DeletedAtIsNil()).
		WithMeterItems().
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainunit.ErrNotFound
		}
		return nil, err
	}
	return toDomainUnitForAccount(record, accountID), nil
}

func (r *UnitRepository) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	record, err := r.client.Unit.Query().
		Where(entunit.ID(id), entunit.DeletedAtIsNil()).
		WithMeterItems().
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainunit.ErrNotFound
		}
		return nil, err
	}
	return toDomainUnit(record), nil
}

func (r *UnitRepository) List(ctx context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	q := r.client.Unit.Query().
		Where(entunit.DeletedAtIsNil()).
		WithMeterItems()
	switch query.Scope {
	case domainunit.ListScopeOwned:
		q = q.Where(entunit.AccountID(query.AccountID))
	case domainunit.ListScopeGranted:
		q = q.Where(r.grantedUnitPredicate(query.AccountID))
	case domainunit.ListScopeAll, "":
		q = q.Where(entunit.Or(
			entunit.AccountID(query.AccountID),
			r.grantedUnitPredicate(query.AccountID),
		))
	default:
		return nil, domainunit.ErrInvalidListScope
	}
	records, err := q.Order(entclient.Desc(entunit.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainUnitsForAccount(records, query.AccountID), nil
}

func (r *UnitRepository) grantedUnitPredicate(accountID int) predicate.Unit {
	return entunit.HasUnitAccountRolesWith(
		entunitaccountrole.AccountID(accountID),
		entunitaccountrole.StatusEQ(entunitaccountrole.StatusActive),
		entunitaccountrole.DeletedAtIsNil(),
		entunitaccountrole.HasRoleWith(
			entrole.IsActive(true),
			entrole.DeletedAtIsNil(),
		),
	)
}

func (r *UnitRepository) Update(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	if err := r.ensureMeterItemsInAccount(ctx, unit.AccountID, unit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}
	affected, err := r.client.Unit.Update().
		Where(entunit.ID(unit.ID), entunit.AccountID(unit.AccountID), entunit.DeletedAtIsNil()).
		SetName(unit.Name).
		SetAddress(unit.Address).
		SetCity(unit.City).
		SetProvince(unit.Province).
		ClearMeterItems().
		AddMeterItemIDs(unit.DefaultMeterItemIDs...).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainunit.ErrNameAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainunit.ErrNotFound
	}
	return r.Get(ctx, unit.AccountID, unit.ID)
}

func (r *UnitRepository) Delete(ctx context.Context, accountID int, id int) error {
	hasRooms, err := r.client.Room.Query().
		Where(entroom.UnitID(id), entroom.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if hasRooms {
		return domainunit.ErrUnitHasRooms
	}
	affected, err := r.client.Unit.Update().
		Where(entunit.ID(id), entunit.AccountID(accountID), entunit.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainunit.ErrNotFound
	}
	return nil
}

func (r *UnitRepository) ensureMeterItemsInAccount(ctx context.Context, accountID int, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	count, err := r.client.MeterItem.Query().
		Where(entmeteritem.IDIn(ids...), entmeteritem.AccountID(accountID), entmeteritem.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(ids) {
		return domainunit.ErrMeterItemNotFound
	}
	return nil
}

func toDomainUnit(record *entclient.Unit) *domainunit.Unit {
	if record == nil {
		return nil
	}
	unit := &domainunit.Unit{
		ID:        record.ID,
		AccountID: record.AccountID,
		Name:      record.Name,
		Address:   record.Address,
		City:      record.City,
		Province:  record.Province,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	if record.Edges.MeterItems != nil {
		unit.DefaultMeterItems = make([]*domainunit.MeterItem, 0, len(record.Edges.MeterItems))
		unit.DefaultMeterItemIDs = make([]int, 0, len(record.Edges.MeterItems))
		for _, item := range record.Edges.MeterItems {
			unit.DefaultMeterItemIDs = append(unit.DefaultMeterItemIDs, item.ID)
			unit.DefaultMeterItems = append(unit.DefaultMeterItems, &domainunit.MeterItem{
				ID:        item.ID,
				Code:      item.Code,
				Unit:      item.Unit,
				UnitPrice: item.UnitPrice,
				IsSystem:  item.IsSystem,
				CalcType:  item.CalcType.String(),
				Remark:    item.Remark,
			})
		}
	}
	return unit
}

func toDomainUnitForAccount(record *entclient.Unit, accountID int) *domainunit.Unit {
	unit := toDomainUnit(record)
	if unit == nil {
		return nil
	}
	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}
	return unit
}

func toDomainUnitsForAccount(records []*entclient.Unit, accountID int) []*domainunit.Unit {
	units := make([]*domainunit.Unit, 0, len(records))
	for _, record := range records {
		units = append(units, toDomainUnitForAccount(record, accountID))
	}
	return units
}

func toDomainUnits(records []*entclient.Unit) []*domainunit.Unit {
	units := make([]*domainunit.Unit, 0, len(records))
	for _, record := range records {
		units = append(units, toDomainUnit(record))
	}
	return units
}
