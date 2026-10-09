package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"

	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
)

// MeterReadingRepository 提供房间抄表记录的 GORM 持久化实现。
type MeterReadingRepository struct {
	q *query.Query
}

// NewMeterReadingRepository 创建抄表记录仓储。
func NewMeterReadingRepository(client *gorm.DB) *MeterReadingRepository {
	return &MeterReadingRepository{q: query.Use(client)}
}

// Create 创建账号名下房间的抄表记录。
func (r *MeterReadingRepository) Create(ctx context.Context, accountID string, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	if err := r.ensureMeterReadingRelationsInAccount(ctx, r.q, accountID, reading.RoomID, reading.MeterItemID); err != nil {
		return nil, err
	}
	record, err := r.createMeterReading(ctx, r.q, reading)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainmeterreading.ErrAlreadyExists
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// Get 获取账号名下指定抄表记录。
func (r *MeterReadingRepository) Get(ctx context.Context, accountID string, id int) (*domainmeterreading.MeterReading, error) {
	record, err := r.q.MeterReading.WithContext(ctx).
		Where(
			r.q.MeterReading.ID.Eq(id),
			r.q.MeterReading.DeletedAt.IsNull(),
			gen.Columns{r.q.MeterReading.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// GetByRoomMeterItemPeriod 获取账号名下房间指定月份的抄表记录。
func (r *MeterReadingRepository) GetByRoomMeterItemPeriod(ctx context.Context, accountID string, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	record, err := r.q.MeterReading.WithContext(ctx).
		Where(
			r.q.MeterReading.RoomID.Eq(roomID),
			r.q.MeterReading.MeterItemID.Eq(meterItemID),
			r.q.MeterReading.PeriodMonth.Eq(periodMonth),
			r.q.MeterReading.DeletedAt.IsNull(),
			gen.Columns{r.q.MeterReading.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// GetLatestBeforePeriod 获取账号名下房间指定抄表项在目标月份之前的最近读数。
func (r *MeterReadingRepository) GetLatestBeforePeriod(ctx context.Context, accountID string, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	record, err := r.q.MeterReading.WithContext(ctx).
		Where(
			r.q.MeterReading.RoomID.Eq(roomID),
			r.q.MeterReading.MeterItemID.Eq(meterItemID),
			r.q.MeterReading.PeriodMonth.Lt(periodMonth),
			r.q.MeterReading.DeletedAt.IsNull(),
			gen.Columns{r.q.MeterReading.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Order(r.q.MeterReading.PeriodMonth.Desc(), r.q.MeterReading.ID.Desc()).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// List 查询账号名下的抄表记录列表。
func (r *MeterReadingRepository) List(ctx context.Context, accountID string, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	q := r.queryMeterReadings(ctx, accountID, query)
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(r.q.MeterReading.PeriodMonth.Desc(), r.q.MeterReading.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainMeterReadings(records), nil
}

// SaveRoomMonth 事务性保存同一房间同一月份的一组抄表记录。
func (r *MeterReadingRepository) SaveRoomMonth(ctx context.Context, accountID string, readings []*domainmeterreading.MeterReading) ([]*domainmeterreading.MeterReading, error) {
	if len(readings) == 0 {
		return nil, domainmeterreading.ErrReadingsRequired
	}

	saved := make([]*domainmeterreading.MeterReading, 0, len(readings))
	err := query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(client *query.Query) error {
		for _, reading := range readings {
			if err := r.ensureMeterReadingRelationsInAccount(ctx, client, accountID, reading.RoomID, reading.MeterItemID); err != nil {
				return err
			}

			record, err := client.MeterReading.WithContext(ctx).
				Where(
					client.MeterReading.RoomID.Eq(reading.RoomID),
					client.MeterReading.MeterItemID.Eq(reading.MeterItemID),
					client.MeterReading.PeriodMonth.Eq(reading.PeriodMonth),
					client.MeterReading.DeletedAt.IsNull(),
					gen.Columns{client.MeterReading.RoomID}.In(accountRooms(ctx, client, accountID).Select(client.Room.ID)),
				).
				Take()
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				created, err := r.createMeterReading(ctx, client, reading)
				if err != nil {
					if isConstraintError(err) {
						return domainmeterreading.ErrAlreadyExists
					}
					return err
				}
				saved = append(saved, toDomainMeterReading(created))
				continue
			}

			err = r.updateMeterReadingRecord(client.MeterReading.WithContext(ctx).Where(client.MeterReading.ID.Eq(record.ID)), reading)
			if err != nil {
				if isConstraintError(err) {
					return domainmeterreading.ErrAlreadyExists
				}
				return err
			}
			updated, err := client.MeterReading.WithContext(ctx).Where(client.MeterReading.ID.Eq(record.ID)).Take()
			if err != nil {
				return err
			}

			saved = append(saved, toDomainMeterReading(updated))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// Update 更新账号名下指定抄表记录。
func (r *MeterReadingRepository) Update(ctx context.Context, accountID string, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	if err := r.ensureMeterReadingRelationsInAccount(ctx, r.q, accountID, reading.RoomID, reading.MeterItemID); err != nil {
		return nil, err
	}
	m := r.q.MeterReading
	q := m.WithContext(ctx).Where(m.ID.Eq(reading.ID), m.DeletedAt.IsNull(),
		gen.Columns{m.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)))
	if err := r.updateMeterReadingRecord(q, reading); err != nil {
		return nil, err
	}
	return r.Get(ctx, accountID, reading.ID)
}

// Delete 软删除账号名下指定抄表记录。
func (r *MeterReadingRepository) Delete(ctx context.Context, accountID string, id int) error {
	result, err := r.q.MeterReading.WithContext(ctx).
		Where(
			r.q.MeterReading.ID.Eq(id),
			r.q.MeterReading.DeletedAt.IsNull(),
			gen.Columns{r.q.MeterReading.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Update(r.q.MeterReading.DeletedAt, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainmeterreading.ErrNotFound
	}
	return nil
}

// queryMeterReadings 构建带账号归属约束的抄表记录查询。
func (r *MeterReadingRepository) queryMeterReadings(ctx context.Context, accountID string, params domainmeterreading.ListQuery) query.IMeterReadingDo {
	rooms := accountRooms(ctx, r.q, accountID)
	if params.UnitID > 0 {
		rooms = rooms.Where(r.q.Room.UnitID.Eq(params.UnitID))
	}
	q := r.q.MeterReading.WithContext(ctx).Where(
		r.q.MeterReading.DeletedAt.IsNull(),
		gen.Columns{r.q.MeterReading.RoomID}.In(rooms.Select(r.q.Room.ID)),
	)

	if params.RoomID > 0 {
		q = q.Where(r.q.MeterReading.RoomID.Eq(params.RoomID))
	}
	if params.MeterItemID > 0 {
		q = q.Where(r.q.MeterReading.MeterItemID.Eq(params.MeterItemID))
	}
	if params.PeriodMonth != "" {
		q = q.Where(r.q.MeterReading.PeriodMonth.Eq(params.PeriodMonth))
	}
	if params.PeriodMonthFrom != "" {
		q = q.Where(r.q.MeterReading.PeriodMonth.Gte(params.PeriodMonthFrom))
	}
	if params.PeriodMonthTo != "" {
		q = q.Where(r.q.MeterReading.PeriodMonth.Lte(params.PeriodMonthTo))
	}
	if params.Source != "" {
		q = q.Where(r.q.MeterReading.Source.Eq(params.Source))
	}
	return q
}

// createMeterReading 创建抄表记录。
func (r *MeterReadingRepository) createMeterReading(ctx context.Context, client *query.Query, reading *domainmeterreading.MeterReading) (*models.MeterReading, error) {
	record := toModelMeterReading(reading)
	if err := client.MeterReading.WithContext(ctx).Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

// updateMeterReadingRecord 更新抄表记录实体。
func (r *MeterReadingRepository) updateMeterReadingRecord(q query.IMeterReadingDo, reading *domainmeterreading.MeterReading) error {
	m := r.q.MeterReading
	result, err := q.Select(
		m.RoomID,
		m.MeterItemID,
		m.PeriodMonth,
		m.PreviousReading,
		m.CurrentReading,
		m.UsageAmount,
		m.ReadingTime,
		m.ReaderID,
		m.Source,
		m.Remark,
	).Updates(toModelMeterReading(reading))
	if err != nil {
		if isConstraintError(err) {
			return domainmeterreading.ErrAlreadyExists
		}
		return err
	}
	if result.RowsAffected == 0 {
		count, err := q.Count()
		if err != nil {
			return err
		}
		if count == 0 {
			return domainmeterreading.ErrNotFound
		}
	}
	return nil
}

func toModelMeterReading(reading *domainmeterreading.MeterReading) *models.MeterReading {
	return &models.MeterReading{
		RoomID:          reading.RoomID,
		MeterItemID:     reading.MeterItemID,
		PeriodMonth:     reading.PeriodMonth,
		PreviousReading: reading.PreviousReading,
		CurrentReading:  reading.CurrentReading,
		UsageAmount:     reading.UsageAmount,
		ReadingTime:     reading.ReadingTime,
		ReaderID:        reading.ReaderID,
		Source:          reading.Source,
		Remark:          reading.Remark,
	}
}

// ensureMeterReadingRelationsInAccount 确认房间属于当前账号且房间已绑定抄表项。
func (r *MeterReadingRepository) ensureMeterReadingRelationsInAccount(ctx context.Context, client *query.Query, accountID string, roomID int, meterItemID int) error {
	count, err := accountRooms(ctx, client, accountID).Where(client.Room.ID.Eq(roomID)).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainmeterreading.ErrRoomNotFound
	}
	m := client.MeterItem
	count, err = m.WithContext(ctx).Where(m.ID.Eq(meterItemID), m.AccountID.Eq(accountID), m.DeletedAt.IsNull()).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainmeterreading.ErrMeterItemNotFound
	}
	binding := client.RoomMeterItem
	count, err = binding.WithContext(ctx).Where(binding.RoomID.Eq(roomID), binding.MeterItemID.Eq(meterItemID)).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainmeterreading.ErrRoomMeterItemNotAssigned
	}
	return nil
}

// toDomainMeterReading 将 GORM 抄表记录转换为领域对象。
func toDomainMeterReading(record *models.MeterReading) *domainmeterreading.MeterReading {
	if record == nil {
		return nil
	}
	return &domainmeterreading.MeterReading{
		ID:              record.ID,
		RoomID:          record.RoomID,
		MeterItemID:     record.MeterItemID,
		PeriodMonth:     record.PeriodMonth,
		PreviousReading: record.PreviousReading,
		CurrentReading:  record.CurrentReading,
		UsageAmount:     record.UsageAmount,
		ReadingTime:     record.ReadingTime,
		ReaderID:        record.ReaderID,
		Source:          record.Source,
		Remark:          record.Remark,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

// toDomainMeterReadings 将 GORM 抄表记录列表转换为领域对象列表。
func toDomainMeterReadings(records []*models.MeterReading) []*domainmeterreading.MeterReading {
	readings := make([]*domainmeterreading.MeterReading, 0, len(records))
	for _, record := range records {
		readings = append(readings, toDomainMeterReading(record))
	}
	return readings
}
