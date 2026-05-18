package repo

import (
	"context"
	"time"

	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	domainroom "github.com/superwhys/rent-manager-server/internal/domain/room"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entmeteritem "github.com/superwhys/rent-manager-server/internal/ent/meteritem"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
	entroomtype "github.com/superwhys/rent-manager-server/internal/ent/roomtype"
	entunit "github.com/superwhys/rent-manager-server/internal/ent/unit"
)

type RoomRepository struct {
	client *entclient.Client
}

func NewRoomRepository(client *entclient.Client) *RoomRepository {
	return &RoomRepository{client: client}
}

func (r *RoomRepository) Create(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if err := r.ensureRoomRelationsInAccount(ctx, accountID, room.UnitID, room.RoomTypeID, room.MeterItemIDs); err != nil {
		return nil, err
	}
	record, err := r.client.Room.Create().
		SetRoomNo(room.RoomNo).
		SetFloor(room.Floor).
		SetPrice(room.Price).
		SetSize(room.Size).
		SetUnitID(room.UnitID).
		SetRoomTypeID(room.RoomTypeID).
		SetIsActive(room.IsActive).
		AddMeterItemIDs(room.MeterItemIDs...).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, accountID, record.ID)
}

func (r *RoomRepository) Get(ctx context.Context, accountID int, id int) (*domainroom.Room, error) {
	record, err := r.client.Room.Query().
		Where(
			entroom.ID(id),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		WithMeterItems().
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainroom.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id int) (*domainroom.Room, error) {
	record, err := r.client.Room.Query().
		Where(entroom.ID(id), entroom.DeletedAtIsNil()).
		WithMeterItems().
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainroom.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) List(ctx context.Context, accountID int, query domainroom.ListQuery) ([]*domainroom.Room, error) {
	q := r.client.Room.Query().
		Where(
			entroom.UnitID(query.UnitID),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		WithMeterItems()
	if query.RoomTypeID > 0 {
		q = q.Where(entroom.RoomTypeID(query.RoomTypeID))
	}
	if query.IsActive != nil {
		q = q.Where(entroom.IsActive(*query.IsActive))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entroom.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRooms(records), nil
}

func (r *RoomRepository) Update(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if err := r.ensureRoomRelationsInAccount(ctx, accountID, room.UnitID, room.RoomTypeID, room.MeterItemIDs); err != nil {
		return nil, err
	}
	affected, err := r.client.Room.Update().
		Where(
			entroom.ID(room.ID),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		SetRoomNo(room.RoomNo).
		SetFloor(room.Floor).
		SetPrice(room.Price).
		SetSize(room.Size).
		SetUnitID(room.UnitID).
		SetRoomTypeID(room.RoomTypeID).
		SetIsActive(room.IsActive).
		ClearMeterItems().
		AddMeterItemIDs(room.MeterItemIDs...).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, domainroom.ErrNotFound
	}
	return r.Get(ctx, accountID, room.ID)
}

func (r *RoomRepository) Delete(ctx context.Context, accountID int, id int) error {
	affected, err := r.client.Room.Update().
		Where(
			entroom.ID(id),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		SetDeletedAt(time.Now()).
		SetIsActive(false).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainroom.ErrNotFound
	}
	return nil
}

func (r *RoomRepository) ensureRoomRelationsInAccount(ctx context.Context, accountID int, unitID int, roomTypeID int, meterItemIDs []int) error {
	unitExists, err := r.client.Unit.Query().
		Where(entunit.ID(unitID), entunit.AccountID(accountID), entunit.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !unitExists {
		return domainroom.ErrUnitRequired
	}
	roomTypeExists, err := r.client.RoomType.Query().
		Where(entroomtype.ID(roomTypeID), entroomtype.AccountID(accountID), entroomtype.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !roomTypeExists {
		return domainroom.ErrRoomTypeRequired
	}
	if len(meterItemIDs) == 0 {
		return nil
	}
	count, err := r.client.MeterItem.Query().
		Where(entmeteritem.IDIn(meterItemIDs...), entmeteritem.AccountID(accountID), entmeteritem.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(meterItemIDs) {
		return domainroom.ErrMeterItemNotFound
	}
	return nil
}

func toDomainRoom(record *entclient.Room) *domainroom.Room {
	if record == nil {
		return nil
	}
	room := &domainroom.Room{
		ID:         record.ID,
		RoomNo:     record.RoomNo,
		Floor:      record.Floor,
		Price:      record.Price,
		Size:       record.Size,
		UnitID:     record.UnitID,
		RoomTypeID: record.RoomTypeID,
		IsActive:   record.IsActive,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
	if record.Edges.MeterItems != nil {
		room.MeterItems = make([]*domainmeteritem.MeterItem, 0, len(record.Edges.MeterItems))
		room.MeterItemIDs = make([]int, 0, len(record.Edges.MeterItems))
		for _, item := range record.Edges.MeterItems {
			room.MeterItemIDs = append(room.MeterItemIDs, item.ID)
			room.MeterItems = append(room.MeterItems, toDomainMeterItem(item))
		}
	}
	return room
}

func toDomainRooms(records []*entclient.Room) []*domainroom.Room {
	rooms := make([]*domainroom.Room, 0, len(records))
	for _, record := range records {
		rooms = append(rooms, toDomainRoom(record))
	}
	return rooms
}
