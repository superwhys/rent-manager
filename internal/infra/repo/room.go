package repo

import (
	"context"

	"github.com/superwhys/rent-manager-server/internal/domain/room"

	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
)

type RoomRepository struct {
	client *entclient.Client
}

func NewRoomRepository(client *entclient.Client) *RoomRepository {
	return &RoomRepository{client: client}
}

func (r *RoomRepository) Create(ctx context.Context, createRoom *room.Room) (*room.Room, error) {
	record, err := r.client.Room.Create().
		SetRoomNo(createRoom.RoomNo).
		SetFloor(createRoom.Floor).
		SetPrice(createRoom.Price).
		SetSize(createRoom.Size).
		SetBuildingID(createRoom.BuildingID).
		SetRoomTypeID(createRoom.RoomTypeID).
		SetIsActive(createRoom.IsActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) Get(ctx context.Context, id int) (*room.Room, error) {
	record, err := r.client.Room.Query().
		Where(entroom.ID(id)).
		Where(entroom.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, mapRoomError(err)
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) List(ctx context.Context, query room.ListQuery) ([]*room.Room, error) {
	builder := r.client.Room.Query().
		Where(entroom.DeletedAtIsNil()).
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

func (r *RoomRepository) Update(ctx context.Context, updateRoom *room.Room) (*room.Room, error) {
	record, err := r.client.Room.UpdateOneID(updateRoom.ID).
		SetRoomNo(updateRoom.RoomNo).
		SetFloor(updateRoom.Floor).
		SetPrice(updateRoom.Price).
		SetSize(updateRoom.Size).
		SetBuildingID(updateRoom.BuildingID).
		SetRoomTypeID(updateRoom.RoomTypeID).
		SetIsActive(updateRoom.IsActive).
		Save(ctx)
	if err != nil {
		return nil, mapRoomError(err)
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) Delete(ctx context.Context, id int) error {
	err := r.client.Room.UpdateOneID(id).
		SetDeletedAt(entroom.DefaultUpdatedAt()).
		Exec(ctx)
	return mapRoomError(err)
}

func toDomainRoom(record *entclient.Room) *room.Room {
	if record == nil {
		return nil
	}
	return &room.Room{
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
