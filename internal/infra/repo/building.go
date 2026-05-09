package repo

import (
	"context"

	domainbuilding "github.com/superwhys/rent-manager-server/internal/domain/building"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entbuilding "github.com/superwhys/rent-manager-server/internal/ent/building"
	entmeteritem "github.com/superwhys/rent-manager-server/internal/ent/meteritem"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
)

type BuildingRepository struct {
	client *entclient.Client
}

func NewBuildingRepository(client *entclient.Client) *BuildingRepository {
	return &BuildingRepository{client: client}
}

func (r *BuildingRepository) Create(ctx context.Context, createBuilding *domainbuilding.Building) (*domainbuilding.Building, error) {
	if err := r.ensureMeterItemsOwned(ctx, createBuilding.AccountID, createBuilding.DefaultMeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Building.Create().
		SetAccountID(createBuilding.AccountID).
		SetName(createBuilding.Name).
		SetAddress(createBuilding.Address).
		SetCity(createBuilding.City).
		SetProvince(createBuilding.Province)
	if len(createBuilding.DefaultMeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(createBuilding.DefaultMeterItemIDs...)
	}
	record, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, createBuilding.AccountID, record.ID)
}

func (r *BuildingRepository) Get(ctx context.Context, accountID int, id int) (*domainbuilding.Building, error) {
	record, err := r.client.Building.Query().
		Where(entbuilding.ID(id)).
		Where(entbuilding.AccountID(accountID)).
		Where(entbuilding.DeletedAtIsNil()).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return nil, mapBuildingError(err)
	}

	return toDomainBuilding(record), nil
}

func (r *BuildingRepository) List(ctx context.Context, query domainbuilding.ListQuery) ([]*domainbuilding.Building, error) {
	builder := r.client.Building.Query().
		Where(entbuilding.AccountID(query.AccountID)).
		Where(entbuilding.DeletedAtIsNil()).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Order(entclient.Desc(entbuilding.FieldID))

	if query.Name != "" {
		builder = builder.Where(entbuilding.NameContainsFold(query.Name))
	}
	if query.City != "" {
		builder = builder.Where(entbuilding.CityContainsFold(query.City))
	}
	if query.Province != "" {
		builder = builder.Where(entbuilding.ProvinceContainsFold(query.Province))
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
	buildings := make([]*domainbuilding.Building, 0, len(records))
	for _, record := range records {
		buildings = append(buildings, toDomainBuilding(record))
	}
	return buildings, nil
}

func (r *BuildingRepository) Update(ctx context.Context, updateBuilding *domainbuilding.Building) (*domainbuilding.Building, error) {
	if _, err := r.Get(ctx, updateBuilding.AccountID, updateBuilding.ID); err != nil {
		return nil, err
	}
	if err := r.ensureMeterItemsOwned(ctx, updateBuilding.AccountID, updateBuilding.DefaultMeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Building.Update().
		Where(entbuilding.ID(updateBuilding.ID)).
		Where(entbuilding.AccountID(updateBuilding.AccountID)).
		Where(entbuilding.DeletedAtIsNil()).
		SetName(updateBuilding.Name).
		SetAddress(updateBuilding.Address).
		SetCity(updateBuilding.City).
		SetProvince(updateBuilding.Province).
		ClearMeterItems()
	if len(updateBuilding.DefaultMeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(updateBuilding.DefaultMeterItemIDs...)
	}
	if _, err := builder.Save(ctx); err != nil {
		return nil, mapBuildingError(err)
	}
	return r.Get(ctx, updateBuilding.AccountID, updateBuilding.ID)
}

func (r *BuildingRepository) Delete(ctx context.Context, accountID int, id int) error {
	if _, err := r.Get(ctx, accountID, id); err != nil {
		return err
	}
	exists, err := r.client.Room.Query().
		Where(entroom.BuildingID(id)).
		Where(entroom.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return domainbuilding.ErrBuildingHasRooms
	}

	affected, err := r.client.Building.Update().
		Where(entbuilding.ID(id)).
		Where(entbuilding.AccountID(accountID)).
		Where(entbuilding.DeletedAtIsNil()).
		SetDeletedAt(entbuilding.DefaultUpdatedAt()).
		Save(ctx)
	if err != nil {
		return mapBuildingError(err)
	}
	if affected == 0 {
		return domainbuilding.ErrNotFound
	}
	return nil
}

func (r *BuildingRepository) ensureMeterItemsOwned(ctx context.Context, accountID int, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	count, err := r.client.MeterItem.Query().
		Where(entmeteritem.IDIn(ids...)).
		Where(entmeteritem.AccountID(accountID)).
		Where(entmeteritem.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(ids) {
		return domainbuilding.ErrMeterItemNotFound
	}
	return nil
}

func toDomainBuilding(record *entclient.Building) *domainbuilding.Building {
	if record == nil {
		return nil
	}
	building := &domainbuilding.Building{
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
		building.DefaultMeterItems = make([]*domainbuilding.MeterItem, 0, len(record.Edges.MeterItems))
		building.DefaultMeterItemIDs = make([]int, 0, len(record.Edges.MeterItems))
		for _, item := range record.Edges.MeterItems {
			building.DefaultMeterItemIDs = append(building.DefaultMeterItemIDs, item.ID)
			building.DefaultMeterItems = append(building.DefaultMeterItems, &domainbuilding.MeterItem{
				ID:       item.ID,
				Code:     item.Code,
				Unit:     item.Unit,
				IsSystem: item.IsSystem,
				CalcType: item.CalcType.String(),
				Remark:   item.Remark,
			})
		}
	}
	return building
}

func mapBuildingError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domainbuilding.ErrNotFound
	}
	return err
}
