package repo

import (
	"context"

	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	"github.com/superwhys/rent-manager-server/internal/domain/room"

	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entbuilding "github.com/superwhys/rent-manager-server/internal/ent/building"
	entmeteritem "github.com/superwhys/rent-manager-server/internal/ent/meteritem"
	"github.com/superwhys/rent-manager-server/internal/ent/predicate"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
)

type RoomRepository struct {
	client *entclient.Client
}

func NewRoomRepository(client *entclient.Client) *RoomRepository {
	return &RoomRepository{client: client}
}

func (r *RoomRepository) Create(ctx context.Context, organizationID int, createRoom *room.Room) (*room.Room, error) {
	if err := r.ensureBuildingOwned(ctx, organizationID, createRoom.BuildingID); err != nil {
		return nil, err
	}
	if err := r.ensureRoomMeterItemsOwned(ctx, organizationID, createRoom.MeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Room.Create().
		SetRoomNo(createRoom.RoomNo).
		SetFloor(createRoom.Floor).
		SetPrice(createRoom.Price).
		SetSize(createRoom.Size).
		SetBuildingID(createRoom.BuildingID).
		SetRoomTypeID(createRoom.RoomTypeID).
		SetIsActive(createRoom.IsActive)
	if len(createRoom.MeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(createRoom.MeterItemIDs...)
	}
	record, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, organizationID, record.ID)
}

func (r *RoomRepository) Get(ctx context.Context, organizationID int, id int) (*room.Room, error) {
	record, err := r.client.Room.Query().
		Where(entroom.ID(id)).
		Where(entroom.DeletedAtIsNil()).
		Where(roomOrganizationPredicate(organizationID)).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return nil, mapRoomError(err)
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) List(ctx context.Context, organizationID int, query room.ListQuery) ([]*room.Room, error) {
	builder := r.client.Room.Query().
		Where(entroom.DeletedAtIsNil()).
		Where(roomOrganizationPredicate(organizationID)).
		WithMeterItems(func(query *entclient.MeterItemQuery) {
			query.Where(entmeteritem.DeletedAtIsNil())
		}).
		Order(entclient.Desc(entroom.FieldID))
	if query.BuildingID > 0 {
		builder = builder.Where(entroom.BuildingID(query.BuildingID))
	}
	if query.RoomTypeID > 0 {
		builder = builder.Where(entroom.RoomTypeID(query.RoomTypeID))
	}
	if query.IsActive != nil {
		builder = builder.Where(entroom.IsActive(*query.IsActive))
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
	rooms := make([]*room.Room, 0, len(records))
	for _, record := range records {
		rooms = append(rooms, toDomainRoom(record))
	}
	return rooms, nil
}

func (r *RoomRepository) Update(ctx context.Context, organizationID int, updateRoom *room.Room) (*room.Room, error) {
	if _, err := r.Get(ctx, organizationID, updateRoom.ID); err != nil {
		return nil, err
	}
	if err := r.ensureBuildingOwned(ctx, organizationID, updateRoom.BuildingID); err != nil {
		return nil, err
	}
	if err := r.ensureRoomMeterItemsOwned(ctx, organizationID, updateRoom.MeterItemIDs); err != nil {
		return nil, err
	}

	builder := r.client.Room.Update().
		Where(entroom.ID(updateRoom.ID)).
		Where(entroom.DeletedAtIsNil()).
		Where(roomOrganizationPredicate(organizationID)).
		SetRoomNo(updateRoom.RoomNo).
		SetFloor(updateRoom.Floor).
		SetPrice(updateRoom.Price).
		SetSize(updateRoom.Size).
		SetBuildingID(updateRoom.BuildingID).
		SetRoomTypeID(updateRoom.RoomTypeID).
		SetIsActive(updateRoom.IsActive).
		ClearMeterItems()
	if len(updateRoom.MeterItemIDs) > 0 {
		builder = builder.AddMeterItemIDs(updateRoom.MeterItemIDs...)
	}
	affected, err := builder.Save(ctx)
	if err != nil {
		return nil, mapRoomError(err)
	}
	if affected == 0 {
		return nil, room.ErrNotFound
	}
	return r.Get(ctx, organizationID, updateRoom.ID)
}

func (r *RoomRepository) Delete(ctx context.Context, organizationID int, id int) error {
	affected, err := r.client.Room.Update().
		Where(entroom.ID(id)).
		Where(entroom.DeletedAtIsNil()).
		Where(roomOrganizationPredicate(organizationID)).
		SetDeletedAt(entroom.DefaultUpdatedAt()).
		Save(ctx)
	if err != nil {
		return mapRoomError(err)
	}
	if affected == 0 {
		return room.ErrNotFound
	}
	return nil
}

func (r *RoomRepository) ensureBuildingOwned(ctx context.Context, organizationID int, id int) error {
	exists, err := r.client.Building.Query().
		Where(entbuilding.ID(id)).
		Where(entbuilding.OrganizationID(organizationID)).
		Where(entbuilding.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return room.ErrNotFound
	}
	return nil
}

func (r *RoomRepository) ensureRoomMeterItemsOwned(ctx context.Context, organizationID int, ids []int) error {
	if err := ensureMeterItemsOwned(ctx, r.client, organizationID, ids); err != nil {
		if err == domainmeteritem.ErrNotFound {
			return room.ErrMeterItemNotFound
		}
		return err
	}
	return nil
}

func roomOrganizationPredicate(organizationID int) predicate.Room {
	return entroom.HasBuildingWith(
		entbuilding.OrganizationID(organizationID),
		entbuilding.DeletedAtIsNil(),
	)
}

func toDomainRoom(record *entclient.Room) *room.Room {
	if record == nil {
		return nil
	}
	domainRoom := &room.Room{
		ID:         record.ID,
		RoomNo:     record.RoomNo,
		Floor:      record.Floor,
		Price:      record.Price,
		Size:       record.Size,
		BuildingID: record.BuildingID,
		RoomTypeID: record.RoomTypeID,
		IsActive:   record.IsActive,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
	if len(record.Edges.MeterItems) > 0 {
		domainRoom.MeterItemIDs = make([]int, 0, len(record.Edges.MeterItems))
		domainRoom.MeterItems = make([]*domainmeteritem.MeterItem, 0, len(record.Edges.MeterItems))
		for _, item := range record.Edges.MeterItems {
			domainRoom.MeterItemIDs = append(domainRoom.MeterItemIDs, item.ID)
			domainRoom.MeterItems = append(domainRoom.MeterItems, &domainmeteritem.MeterItem{
				ID:             item.ID,
				OrganizationID: item.OrganizationID,
				Code:           item.Code,
				Unit:           item.Unit,
				IsSystem:       item.IsSystem,
				CalcType:       item.CalcType.String(),
				Remark:         item.Remark,
				CreatedAt:      item.CreatedAt,
				UpdatedAt:      item.UpdatedAt,
			})
		}
	}
	return domainRoom
}

func mapRoomError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return room.ErrNotFound
	}
	return err
}
