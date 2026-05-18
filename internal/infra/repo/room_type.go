package repo

import (
	"context"
	"time"

	domainroomtype "github.com/superwhys/rent-manager-server/internal/domain/roomtype"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entroom "github.com/superwhys/rent-manager-server/internal/ent/room"
	entroomtype "github.com/superwhys/rent-manager-server/internal/ent/roomtype"
)

type RoomTypeRepository struct {
	client *entclient.Client
}

func NewRoomTypeRepository(client *entclient.Client) *RoomTypeRepository {
	return &RoomTypeRepository{client: client}
}

func (r *RoomTypeRepository) Create(ctx context.Context, roomType *domainroomtype.RoomType) (*domainroomtype.RoomType, error) {
	record, err := r.client.RoomType.Create().
		SetAccountID(roomType.AccountID).
		SetName(roomType.Name).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainroomtype.ErrNameAlreadyExists
		}
		return nil, err
	}
	return toDomainRoomType(record), nil
}

func (r *RoomTypeRepository) Get(ctx context.Context, accountID int, id int) (*domainroomtype.RoomType, error) {
	record, err := r.client.RoomType.Query().
		Where(entroomtype.ID(id), entroomtype.AccountID(accountID), entroomtype.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainroomtype.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoomType(record), nil
}

func (r *RoomTypeRepository) List(ctx context.Context, query domainroomtype.ListQuery) ([]*domainroomtype.RoomType, error) {
	q := r.client.RoomType.Query().
		Where(entroomtype.AccountID(query.AccountID), entroomtype.DeletedAtIsNil())
	if query.Name != "" {
		q = q.Where(entroomtype.NameContains(query.Name))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entroomtype.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRoomTypes(records), nil
}

func (r *RoomTypeRepository) Update(ctx context.Context, roomType *domainroomtype.RoomType) (*domainroomtype.RoomType, error) {
	affected, err := r.client.RoomType.Update().
		Where(entroomtype.ID(roomType.ID), entroomtype.AccountID(roomType.AccountID), entroomtype.DeletedAtIsNil()).
		SetName(roomType.Name).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainroomtype.ErrNameAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainroomtype.ErrNotFound
	}
	return r.Get(ctx, roomType.AccountID, roomType.ID)
}

func (r *RoomTypeRepository) Delete(ctx context.Context, accountID int, id int) error {
	hasRooms, err := r.client.Room.Query().
		Where(entroom.RoomTypeID(id), entroom.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if hasRooms {
		return domainroomtype.ErrRoomTypeHasRooms
	}
	affected, err := r.client.RoomType.Update().
		Where(entroomtype.ID(id), entroomtype.AccountID(accountID), entroomtype.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainroomtype.ErrNotFound
	}
	return nil
}

func toDomainRoomType(record *entclient.RoomType) *domainroomtype.RoomType {
	if record == nil {
		return nil
	}
	return &domainroomtype.RoomType{
		ID:        record.ID,
		AccountID: record.AccountID,
		Name:      record.Name,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}

func toDomainRoomTypes(records []*entclient.RoomType) []*domainroomtype.RoomType {
	roomTypes := make([]*domainroomtype.RoomType, 0, len(records))
	for _, record := range records {
		roomTypes = append(roomTypes, toDomainRoomType(record))
	}
	return roomTypes
}
