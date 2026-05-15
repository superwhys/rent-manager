package repo

import (
	"context"

	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entmeteritem "github.com/superwhys/rent-manager-server/internal/ent/meteritem"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
	entunit "github.com/superwhys/rent-manager-server/internal/ent/unit"
)

type UnitRepository struct {
	client *entclient.Client
}

func NewUnitRepository(client *entclient.Client) *UnitRepository {
	return &UnitRepository{client: client}
}

func (r *UnitRepository) Create(ctx context.Context, createUnit *domainunit.Unit) (*domainunit.Unit, error) {
	if err := r.ensureMeterItemsOwned(ctx, createUnit.AccountID, createUnit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Unit.Create().
		SetAccountID(createUnit.AccountID).
		SetName(createUnit.Name).
		SetAddress(createUnit.Address).
		SetCity(createUnit.City).
		SetProvince(createUnit.Province)
	if len(createUnit.DefaultMeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(createUnit.DefaultMeterItemIDs...)
	}
	record, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, createUnit.AccountID, record.ID)
}

func (r *UnitRepository) Get(ctx context.Context, accountID int, id int) (*domainunit.Unit, error) {
	record, err := r.client.Unit.Query().
		Where(entunit.ID(id)).
		Where(entunit.AccountID(accountID)).
		Where(entunit.DeletedAtIsNil()).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return nil, mapUnitError(err)
	}

	return toDomainUnit(record), nil
}

func (r *UnitRepository) List(ctx context.Context, query domainunit.ListQuery) ([]*domainunit.Unit, error) {
	builder := r.client.Unit.Query().
		Where(entunit.AccountID(query.AccountID)).
		Where(entunit.DeletedAtIsNil()).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Order(entclient.Desc(entunit.FieldID))

	if query.Name != "" {
		builder = builder.Where(entunit.NameContainsFold(query.Name))
	}
	if query.City != "" {
		builder = builder.Where(entunit.CityContainsFold(query.City))
	}
	if query.Province != "" {
		builder = builder.Where(entunit.ProvinceContainsFold(query.Province))
	}
	if query.Limit > 0 {
		builder = builder.Limit(query.Limit)
	}
	if query.Offset > 0 {
		builder = builder.Offset(query.Offset)
	}
	records, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}
	units := make([]*domainunit.Unit, 0, len(records))
	for _, record := range records {
		units = append(units, toDomainUnit(record))
	}
	return units, nil
}

func (r *UnitRepository) Update(ctx context.Context, updateUnit *domainunit.Unit) (*domainunit.Unit, error) {
	if _, err := r.Get(ctx, updateUnit.AccountID, updateUnit.ID); err != nil {
		return nil, err
	}
	if err := r.ensureMeterItemsOwned(ctx, updateUnit.AccountID, updateUnit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Unit.Update().
		Where(entunit.ID(updateUnit.ID)).
		Where(entunit.AccountID(updateUnit.AccountID)).
		Where(entunit.DeletedAtIsNil()).
		SetName(updateUnit.Name).
		SetAddress(updateUnit.Address).
		SetCity(updateUnit.City).
		SetProvince(updateUnit.Province).
		ClearMeterItems()
	if len(updateUnit.DefaultMeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(updateUnit.DefaultMeterItemIDs...)
	}
	if _, err := builder.Save(ctx); err != nil {
		return nil, mapUnitError(err)
	}
	return r.Get(ctx, updateUnit.AccountID, updateUnit.ID)
}

func (r *UnitRepository) Delete(ctx context.Context, accountID int, id int) error {
	if _, err := r.Get(ctx, accountID, id); err != nil {
		return err
	}
	exists, err := r.client.Room.Query().
		Where(entroom.UnitID(id)).
		Where(entroom.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return domainunit.ErrUnitHasRooms
	}

	affected, err := r.client.Unit.Update().
		Where(entunit.ID(id)).
		Where(entunit.AccountID(accountID)).
		Where(entunit.DeletedAtIsNil()).
		SetDeletedAt(entunit.DefaultUpdatedAt()).
		Save(ctx)
	if err != nil {
		return mapUnitError(err)
	}
	if affected == 0 {
		return domainunit.ErrNotFound
	}
	return nil
}

func (r *UnitRepository) ensureMeterItemsOwned(ctx context.Context, accountID int, ids []int) error {
	if err := ensureMeterItemsOwned(ctx, r.client, accountID, ids); err != nil {
		if err == domainmeteritem.ErrNotFound {
			return domainunit.ErrMeterItemNotFound
		}
		return err
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
	if len(record.Edges.MeterItems) > 0 {
		unit.DefaultMeterItems = make([]*domainunit.MeterItem, 0, len(record.Edges.MeterItems))
		unit.DefaultMeterItemIDs = make([]int, 0, len(record.Edges.MeterItems))
		for _, item := range record.Edges.MeterItems {
			unit.DefaultMeterItemIDs = append(unit.DefaultMeterItemIDs, item.ID)
			unit.DefaultMeterItems = append(unit.DefaultMeterItems, &domainunit.MeterItem{
				ID:       item.ID,
				Code:     item.Code,
				Unit:     item.Unit,
				IsSystem: item.IsSystem,
				CalcType: item.CalcType.String(),
				Remark:   item.Remark,
			})
		}
	}
	return unit
}

func mapUnitError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domainunit.ErrNotFound
	}
	return err
}
