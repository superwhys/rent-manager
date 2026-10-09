package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type RoomRepository struct {
	q *query.Query
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{q: query.Use(db)}
}

func (r *RoomRepository) Create(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if err := r.ensureRoomRelationsInAccount(ctx, accountID, room.UnitID, room.RoomTypeID, room.MeterItemIDs); err != nil {
		return nil, err
	}
	if err := validateByteLength(255, room.RoomNo); err != nil {
		return nil, err
	}
	record := &models.Room{
		RoomNo:     room.RoomNo,
		Floor:      room.Floor,
		Price:      room.Price,
		Size:       room.Size,
		UnitID:     room.UnitID,
		RoomTypeID: room.RoomTypeID,
		IsActive:   &room.IsActive,
	}
	err := query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		if err := tx.Room.WithContext(ctx).Create(record); err != nil {
			return err
		}
		return replaceRoomMeterItems(ctx, tx, record.ID, room.MeterItemIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, accountID, record.ID)
}

func (r *RoomRepository) Get(ctx context.Context, accountID int, id int) (*domainroom.Room, error) {
	room := r.q.Room
	record, err := room.WithContext(ctx).
		Where(room.ID.Eq(id), room.DeletedAt.IsNull(), r.roomAccountPredicate(ctx, accountID)).
		Preload(room.MeterItems).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainroom.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id int) (*domainroom.Room, error) {
	room := r.q.Room
	record, err := room.WithContext(ctx).
		Where(room.ID.Eq(id), room.DeletedAt.IsNull()).
		Preload(room.MeterItems).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainroom.ErrNotFound
		}
		return nil, err
	}
	return toDomainRoom(record), nil
}

func (r *RoomRepository) List(ctx context.Context, accountID int, params domainroom.ListQuery) ([]*domainroom.Room, error) {
	room := r.q.Room
	q := r.queryRooms(ctx, accountID, params).Preload(room.MeterItems)
	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}
	records, err := q.Order(room.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainRooms(records), nil
}

func (r *RoomRepository) Count(ctx context.Context, accountID int, params domainroom.ListQuery) (int, error) {
	count, err := r.queryRooms(ctx, accountID, params).Count()
	return int(count), err
}

func (r *RoomRepository) queryRooms(ctx context.Context, accountID int, params domainroom.ListQuery) query.IRoomDo {
	room := r.q.Room
	q := room.WithContext(ctx).
		Where(room.UnitID.Eq(params.UnitID), room.DeletedAt.IsNull(), r.roomAccountPredicate(ctx, accountID))
	if params.RoomTypeID > 0 {
		q = q.Where(room.RoomTypeID.Eq(params.RoomTypeID))
	}
	if params.IsActive != nil {
		q = q.Where(room.IsActive.Is(*params.IsActive))
	}
	return q
}

func (r *RoomRepository) roomAccountPredicate(ctx context.Context, accountID int) gen.Condition {
	unit := r.q.Unit
	return gen.Exists(unit.WithContext(ctx).Where(
		unit.ID.EqCol(r.q.Room.UnitID),
		unit.AccountID.Eq(accountID),
		unit.DeletedAt.IsNull(),
	))
}

func (r *RoomRepository) Update(ctx context.Context, accountID int, room *domainroom.Room) (*domainroom.Room, error) {
	if err := r.ensureRoomRelationsInAccount(ctx, accountID, room.UnitID, room.RoomTypeID, room.MeterItemIDs); err != nil {
		return nil, err
	}
	if err := validateByteLength(255, room.RoomNo); err != nil {
		return nil, err
	}
	err := query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		repo := &RoomRepository{q: tx}
		r := tx.Room
		target := r.WithContext(ctx).
			Where(r.ID.Eq(room.ID), r.DeletedAt.IsNull(), repo.roomAccountPredicate(ctx, accountID))
		if _, err := target.Select(r.ID).Clauses(clause.Locking{Strength: "UPDATE"}).Take(); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainroom.ErrNotFound
			}
			return err
		}
		_, err := r.WithContext(ctx).
			Where(r.ID.Eq(room.ID), r.DeletedAt.IsNull(), repo.roomAccountPredicate(ctx, accountID)).
			UpdateSimple(
				r.RoomNo.Value(room.RoomNo),
				r.Floor.Value(room.Floor),
				r.Price.Value(room.Price),
				r.Size.Value(room.Size),
				r.UnitID.Value(room.UnitID),
				r.RoomTypeID.Value(room.RoomTypeID),
				r.IsActive.Value(room.IsActive),
			)
		if err != nil {
			return err
		}
		return replaceRoomMeterItems(ctx, tx, room.ID, room.MeterItemIDs)
	})
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, accountID, room.ID)
}

func (r *RoomRepository) Delete(ctx context.Context, accountID int, id int) error {
	room := r.q.Room
	result, err := room.WithContext(ctx).
		Where(room.ID.Eq(id), room.DeletedAt.IsNull(), r.roomAccountPredicate(ctx, accountID)).
		UpdateSimple(room.DeletedAt.Value(time.Now()), room.IsActive.Value(false))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainroom.ErrNotFound
	}
	return nil
}

func (r *RoomRepository) ensureRoomRelationsInAccount(ctx context.Context, accountID int, unitID int, roomTypeID int, meterItemIDs []int) error {
	unit := r.q.Unit
	count, err := unit.WithContext(ctx).
		Where(unit.ID.Eq(unitID), unit.AccountID.Eq(accountID), unit.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainroom.ErrUnitRequired
	}
	roomType := r.q.RoomType
	count, err = roomType.WithContext(ctx).
		Where(roomType.ID.Eq(roomTypeID), roomType.AccountID.Eq(accountID), roomType.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainroom.ErrRoomTypeRequired
	}
	if len(meterItemIDs) == 0 {
		return nil
	}
	item := r.q.MeterItem
	count, err = item.WithContext(ctx).
		Where(item.ID.In(meterItemIDs...), item.AccountID.Eq(accountID), item.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if count != int64(len(meterItemIDs)) {
		return domainroom.ErrMeterItemNotFound
	}
	return nil
}

func replaceRoomMeterItems(ctx context.Context, q *query.Query, roomID int, ids []int) error {
	item := q.RoomMeterItem
	if _, err := item.WithContext(ctx).Where(item.RoomID.Eq(roomID)).Delete(); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	records := make([]*models.RoomMeterItem, 0, len(ids))
	for _, id := range ids {
		records = append(records, &models.RoomMeterItem{RoomID: roomID, MeterItemID: id})
	}
	return item.WithContext(ctx).Create(records...)
}

func toDomainRoom(record *models.Room) *domainroom.Room {
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
		IsActive:   *record.IsActive,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
	}
	if record.MeterItems != nil {
		room.MeterItems = make([]*domainmeteritem.MeterItem, 0, len(record.MeterItems))
		room.MeterItemIDs = make([]int, 0, len(record.MeterItems))
		for _, item := range record.MeterItems {
			room.MeterItemIDs = append(room.MeterItemIDs, item.ID)
			room.MeterItems = append(room.MeterItems, toDomainMeterItem(item))
		}
	}
	return room
}

func toDomainRooms(records []*models.Room) []*domainroom.Room {
	rooms := make([]*domainroom.Room, 0, len(records))
	for _, record := range records {
		rooms = append(rooms, toDomainRoom(record))
	}
	return rooms
}
