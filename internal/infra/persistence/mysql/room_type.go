package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type RoomTypeRepository struct {
	q *query.Query
}

func NewRoomTypeRepository(client *gorm.DB) *RoomTypeRepository {
	return &RoomTypeRepository{q: query.Use(client)}
}

func (r *RoomTypeRepository) Create(
	ctx context.Context,
	roomType *domainroomtype.RoomType,
) (*domainroomtype.RoomType, error) {
	if err := validateByteLength(12, roomType.Name); err != nil {
		return nil, err
	}
	record := &models.RoomType{AccountID: roomType.AccountID, Name: roomType.Name}
	err := r.q.RoomType.WithContext(ctx).Create(record)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainroomtype.ErrNameAlreadyExists
		}
		return nil, err
	}
	return toDomainRoomType(record), nil
}

func (r *RoomTypeRepository) Get(
	ctx context.Context,
	accountID string,
	id int,
) (*domainroomtype.RoomType, error) {
	record, err := r.q.RoomType.WithContext(ctx).
		Where(r.q.RoomType.ID.Eq(id), r.q.RoomType.AccountID.Eq(accountID), r.q.RoomType.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainroomtype.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoomType(record), nil
}

func (r *RoomTypeRepository) List(
	ctx context.Context,
	query domainroomtype.ListQuery,
) ([]*domainroomtype.RoomType, error) {
	q := r.q.RoomType.WithContext(ctx).
		Where(r.q.RoomType.AccountID.Eq(query.AccountID), r.q.RoomType.DeletedAt.IsNull())
	if query.Name != "" {
		q = q.Where(contains(r.q.RoomType.Name, query.Name))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(r.q.RoomType.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainRoomTypes(records), nil
}

func (r *RoomTypeRepository) Update(
	ctx context.Context,
	roomType *domainroomtype.RoomType,
) (*domainroomtype.RoomType, error) {
	if err := validateByteLength(12, roomType.Name); err != nil {
		return nil, err
	}
	result, err := r.q.RoomType.WithContext(ctx).
		Where(r.q.RoomType.ID.Eq(roomType.ID), r.q.RoomType.AccountID.Eq(roomType.AccountID), r.q.RoomType.DeletedAt.IsNull()).
		Update(r.q.RoomType.Name, roomType.Name)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainroomtype.ErrNameAlreadyExists
		}
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, domainroomtype.ErrNotFound
	}
	return r.Get(ctx, roomType.AccountID, roomType.ID)
}

func (r *RoomTypeRepository) Delete(ctx context.Context, accountID string, id int) error {
	roomCount, err := r.q.Room.WithContext(ctx).
		Where(r.q.Room.RoomTypeID.Eq(id), r.q.Room.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if roomCount > 0 {
		return domainroomtype.ErrRoomTypeHasRooms
	}
	result, err := r.q.RoomType.WithContext(ctx).
		Where(r.q.RoomType.ID.Eq(id), r.q.RoomType.AccountID.Eq(accountID), r.q.RoomType.DeletedAt.IsNull()).
		Update(r.q.RoomType.DeletedAt, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainroomtype.ErrNotFound
	}
	return nil
}

func toDomainRoomType(record *models.RoomType) *domainroomtype.RoomType {
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

func toDomainRoomTypes(records []*models.RoomType) []*domainroomtype.RoomType {
	roomTypes := make([]*domainroomtype.RoomType, 0, len(records))
	for _, record := range records {
		roomTypes = append(roomTypes, toDomainRoomType(record))
	}
	return roomTypes
}
