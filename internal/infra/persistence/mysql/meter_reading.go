package mysql

import (
	"context"
	"time"

	"github.com/superwhys/rent-manager/internal/infra/persistence/ent/predicate"

	domainmeterreading "github.com/superwhys/rent-manager/internal/domain/meterreading"
	entclient "github.com/superwhys/rent-manager/internal/infra/persistence/ent"
	entmeteritem "github.com/superwhys/rent-manager/internal/infra/persistence/ent/meteritem"
	entmeterreading "github.com/superwhys/rent-manager/internal/infra/persistence/ent/meterreading"
	entroom "github.com/superwhys/rent-manager/internal/infra/persistence/ent/room"
	entunit "github.com/superwhys/rent-manager/internal/infra/persistence/ent/unit"
)

// MeterReadingRepository 提供房间抄表记录的 Ent 持久化实现。
type MeterReadingRepository struct {
	client *entclient.Client
}

// NewMeterReadingRepository 创建抄表记录仓储。
func NewMeterReadingRepository(client *entclient.Client) *MeterReadingRepository {
	return &MeterReadingRepository{client: client}
}

// Create 创建账号名下房间的抄表记录。
func (r *MeterReadingRepository) Create(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	if err := r.ensureMeterReadingRelationsInAccount(ctx, r.client, accountID, reading.RoomID, reading.MeterItemID); err != nil {
		return nil, err
	}
	record, err := r.createMeterReading(ctx, r.client, reading)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainmeterreading.ErrAlreadyExists
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// Get 获取账号名下指定抄表记录。
func (r *MeterReadingRepository) Get(ctx context.Context, accountID int, id int) (*domainmeterreading.MeterReading, error) {
	record, err := r.client.MeterReading.Query().
		Where(
			entmeterreading.ID(id),
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// GetByRoomMeterItemPeriod 获取账号名下房间指定月份的抄表记录。
func (r *MeterReadingRepository) GetByRoomMeterItemPeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	record, err := r.client.MeterReading.Query().
		Where(
			entmeterreading.RoomID(roomID),
			entmeterreading.MeterItemID(meterItemID),
			entmeterreading.PeriodMonth(periodMonth),
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// GetLatestBeforePeriod 获取账号名下房间指定抄表项在目标月份之前的最近读数。
func (r *MeterReadingRepository) GetLatestBeforePeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*domainmeterreading.MeterReading, error) {
	record, err := r.client.MeterReading.Query().
		Where(
			entmeterreading.RoomID(roomID),
			entmeterreading.MeterItemID(meterItemID),
			entmeterreading.PeriodMonthLT(periodMonth),
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		Order(entclient.Desc(entmeterreading.FieldPeriodMonth), entclient.Desc(entmeterreading.FieldID)).
		First(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainmeterreading.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterReading(record), nil
}

// List 查询账号名下的抄表记录列表。
func (r *MeterReadingRepository) List(ctx context.Context, accountID int, query domainmeterreading.ListQuery) ([]*domainmeterreading.MeterReading, error) {
	q := r.queryMeterReadings(accountID, query)
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entmeterreading.FieldPeriodMonth), entclient.Desc(entmeterreading.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainMeterReadings(records), nil
}

// SaveRoomMonth 事务性保存同一房间同一月份的一组抄表记录。
func (r *MeterReadingRepository) SaveRoomMonth(ctx context.Context, accountID int, readings []*domainmeterreading.MeterReading) ([]*domainmeterreading.MeterReading, error) {
	if len(readings) == 0 {
		return nil, domainmeterreading.ErrReadingsRequired
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	client := tx.Client()
	saved := make([]*domainmeterreading.MeterReading, 0, len(readings))
	for _, reading := range readings {
		if err := r.ensureMeterReadingRelationsInAccount(ctx, client, accountID, reading.RoomID, reading.MeterItemID); err != nil {
			return nil, err
		}

		record, err := client.MeterReading.Query().
			Where(
				entmeterreading.RoomID(reading.RoomID),
				entmeterreading.MeterItemID(reading.MeterItemID),
				entmeterreading.PeriodMonth(reading.PeriodMonth),
				entmeterreading.DeletedAtIsNil(),
				entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
			).
			Only(ctx)
		if err != nil {
			if !entclient.IsNotFound(err) {
				return nil, err
			}
			created, err := r.createMeterReading(ctx, client, reading)
			if err != nil {
				if entclient.IsConstraintError(err) {
					return nil, domainmeterreading.ErrAlreadyExists
				}
				return nil, err
			}
			saved = append(saved, toDomainMeterReading(created))
			continue
		}

		updated, err := r.updateMeterReadingRecord(ctx, client, record.ID, reading)
		if err != nil {
			if entclient.IsConstraintError(err) {
				return nil, domainmeterreading.ErrAlreadyExists
			}
			return nil, err
		}
		saved = append(saved, toDomainMeterReading(updated))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return saved, nil
}

// Update 更新账号名下指定抄表记录。
func (r *MeterReadingRepository) Update(ctx context.Context, accountID int, reading *domainmeterreading.MeterReading) (*domainmeterreading.MeterReading, error) {
	if err := r.ensureMeterReadingRelationsInAccount(ctx, r.client, accountID, reading.RoomID, reading.MeterItemID); err != nil {
		return nil, err
	}
	updater := r.client.MeterReading.Update().
		Where(
			entmeterreading.ID(reading.ID),
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		SetRoomID(reading.RoomID).
		SetMeterItemID(reading.MeterItemID).
		SetPeriodMonth(reading.PeriodMonth).
		SetCurrentReading(reading.CurrentReading).
		SetReadingTime(reading.ReadingTime).
		SetSource(entmeterreading.Source(reading.Source))
	if reading.PreviousReading == nil {
		updater = updater.ClearPreviousReading()
	} else {
		updater = updater.SetPreviousReading(*reading.PreviousReading)
	}
	if reading.UsageAmount == nil {
		updater = updater.ClearUsageAmount()
	} else {
		updater = updater.SetUsageAmount(*reading.UsageAmount)
	}
	if reading.ReaderID == nil {
		updater = updater.ClearReaderID()
	} else {
		updater = updater.SetReaderID(*reading.ReaderID)
	}
	if reading.Remark == nil {
		updater = updater.ClearRemark()
	} else {
		updater = updater.SetRemark(*reading.Remark)
	}
	affected, err := updater.Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainmeterreading.ErrAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainmeterreading.ErrNotFound
	}
	return r.Get(ctx, accountID, reading.ID)
}

// Delete 软删除账号名下指定抄表记录。
func (r *MeterReadingRepository) Delete(ctx context.Context, accountID int, id int) error {
	affected, err := r.client.MeterReading.Update().
		Where(
			entmeterreading.ID(id),
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainmeterreading.ErrNotFound
	}
	return nil
}

// queryMeterReadings 构建带账号归属约束的抄表记录查询。
func (r *MeterReadingRepository) queryMeterReadings(accountID int, query domainmeterreading.ListQuery) *entclient.MeterReadingQuery {
	roomPredicates := r.accountRoomPredicates(accountID)
	if query.UnitID > 0 {
		roomPredicates = append(roomPredicates, entroom.UnitID(query.UnitID))
	}
	q := r.client.MeterReading.Query().
		Where(
			entmeterreading.DeletedAtIsNil(),
			entmeterreading.HasRoomWith(roomPredicates...),
		)
	if query.RoomID > 0 {
		q = q.Where(entmeterreading.RoomID(query.RoomID))
	}
	if query.MeterItemID > 0 {
		q = q.Where(entmeterreading.MeterItemID(query.MeterItemID))
	}
	if query.PeriodMonth != "" {
		q = q.Where(entmeterreading.PeriodMonth(query.PeriodMonth))
	}
	if query.PeriodMonthFrom != "" {
		q = q.Where(entmeterreading.PeriodMonthGTE(query.PeriodMonthFrom))
	}
	if query.PeriodMonthTo != "" {
		q = q.Where(entmeterreading.PeriodMonthLTE(query.PeriodMonthTo))
	}
	if query.Source != "" {
		q = q.Where(entmeterreading.SourceEQ(entmeterreading.Source(query.Source)))
	}
	return q
}

// createMeterReading 创建抄表记录。
func (r *MeterReadingRepository) createMeterReading(ctx context.Context, client *entclient.Client, reading *domainmeterreading.MeterReading) (*entclient.MeterReading, error) {
	return client.MeterReading.Create().
		SetRoomID(reading.RoomID).
		SetMeterItemID(reading.MeterItemID).
		SetPeriodMonth(reading.PeriodMonth).
		SetNillablePreviousReading(reading.PreviousReading).
		SetCurrentReading(reading.CurrentReading).
		SetNillableUsageAmount(reading.UsageAmount).
		SetReadingTime(reading.ReadingTime).
		SetNillableReaderID(reading.ReaderID).
		SetSource(entmeterreading.Source(reading.Source)).
		SetNillableRemark(reading.Remark).
		Save(ctx)
}

// updateMeterReadingRecord 更新抄表记录实体。
func (r *MeterReadingRepository) updateMeterReadingRecord(ctx context.Context, client *entclient.Client, id int, reading *domainmeterreading.MeterReading) (*entclient.MeterReading, error) {
	updater := client.MeterReading.UpdateOneID(id).
		SetRoomID(reading.RoomID).
		SetMeterItemID(reading.MeterItemID).
		SetPeriodMonth(reading.PeriodMonth).
		SetCurrentReading(reading.CurrentReading).
		SetReadingTime(reading.ReadingTime).
		SetSource(entmeterreading.Source(reading.Source))
	if reading.PreviousReading == nil {
		updater = updater.ClearPreviousReading()
	} else {
		updater = updater.SetPreviousReading(*reading.PreviousReading)
	}
	if reading.UsageAmount == nil {
		updater = updater.ClearUsageAmount()
	} else {
		updater = updater.SetUsageAmount(*reading.UsageAmount)
	}
	if reading.ReaderID == nil {
		updater = updater.ClearReaderID()
	} else {
		updater = updater.SetReaderID(*reading.ReaderID)
	}
	if reading.Remark == nil {
		updater = updater.ClearRemark()
	} else {
		updater = updater.SetRemark(*reading.Remark)
	}
	return updater.Save(ctx)
}

// accountRoomPredicates 返回账号名下未删除房间的查询条件。
func (r *MeterReadingRepository) accountRoomPredicates(accountID int) []predicate.Room {
	return []predicate.Room{
		entroom.DeletedAtIsNil(),
		entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
	}
}

// ensureMeterReadingRelationsInAccount 确认房间属于当前账号且房间已绑定抄表项。
func (r *MeterReadingRepository) ensureMeterReadingRelationsInAccount(ctx context.Context, client *entclient.Client, accountID int, roomID int, meterItemID int) error {
	roomExists, err := client.Room.Query().
		Where(
			entroom.ID(roomID),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !roomExists {
		return domainmeterreading.ErrRoomNotFound
	}

	meterItemExists, err := client.MeterItem.Query().
		Where(entmeteritem.ID(meterItemID), entmeteritem.AccountID(accountID), entmeteritem.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !meterItemExists {
		return domainmeterreading.ErrMeterItemNotFound
	}

	assigned, err := client.Room.Query().
		Where(
			entroom.ID(roomID),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
			entroom.HasMeterItemsWith(entmeteritem.ID(meterItemID), entmeteritem.DeletedAtIsNil()),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !assigned {
		return domainmeterreading.ErrRoomMeterItemNotAssigned
	}
	return nil
}

// toDomainMeterReading 将 Ent 抄表记录转换为领域对象。
func toDomainMeterReading(record *entclient.MeterReading) *domainmeterreading.MeterReading {
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
		Source:          record.Source.String(),
		Remark:          record.Remark,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

// toDomainMeterReadings 将 Ent 抄表记录列表转换为领域对象列表。
func toDomainMeterReadings(records []*entclient.MeterReading) []*domainmeterreading.MeterReading {
	readings := make([]*domainmeterreading.MeterReading, 0, len(records))
	for _, record := range records {
		readings = append(readings, toDomainMeterReading(record))
	}
	return readings
}
