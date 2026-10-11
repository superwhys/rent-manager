package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

// RentBillRepository 提供租金账单的 GORM 持久化实现。
type RentBillRepository struct {
	q *query.Query
}

// NewRentBillRepository 创建租金账单仓储。
func NewRentBillRepository(client *gorm.DB) *RentBillRepository {
	return &RentBillRepository{q: query.Use(client)}
}

// Create 创建账号名下房间的租金账单。
func (r *RentBillRepository) Create(
	ctx context.Context,
	accountID string,
	bill *domainrentbill.RentBill,
) (*domainrentbill.RentBill, error) {
	if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
		return nil, err
	}
	record, err := r.createRentBillRecord(ctx, r.q, bill)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainrentbill.ErrAlreadyExists
		}
		return nil, err
	}
	return r.Get(ctx, accountID, record.ID)
}

// Get 获取账号名下指定租金账单。
func (r *RentBillRepository) Get(
	ctx context.Context,
	accountID string,
	id int,
) (*domainrentbill.RentBill, error) {
	record, err := r.q.RentBill.WithContext(ctx).
		Where(
			r.q.RentBill.ID.Eq(id),
			r.q.RentBill.DeletedAt.IsNull(),
			gen.Columns{
				r.q.RentBill.RoomID,
			}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record)
}

// GetByID 获取指定租金账单，不做账号归属过滤。
func (r *RentBillRepository) GetByID(
	ctx context.Context,
	id int,
) (*domainrentbill.RentBill, error) {
	record, err := r.q.RentBill.WithContext(ctx).
		Where(r.q.RentBill.ID.Eq(id), r.q.RentBill.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record)
}

// GetByRoomPeriod 获取账号名下房间指定月份的租金账单。
func (r *RentBillRepository) GetByRoomPeriod(
	ctx context.Context,
	accountID string,
	roomID int,
	periodMonth string,
) (*domainrentbill.RentBill, error) {
	record, err := r.q.RentBill.WithContext(ctx).
		Where(
			r.q.RentBill.RoomID.Eq(roomID),
			r.q.RentBill.PeriodMonth.Eq(periodMonth),
			r.q.RentBill.DeletedAt.IsNull(),
			gen.Columns{
				r.q.RentBill.RoomID,
			}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record)
}

// List 查询账号名下的租金账单列表。
func (r *RentBillRepository) List(
	ctx context.Context,
	accountID string,
	query domainrentbill.ListQuery,
) ([]*domainrentbill.RentBill, error) {
	q := r.queryRentBills(ctx, accountID, query)
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(r.q.RentBill.PeriodMonth.Desc(), r.q.RentBill.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainRentBills(records)
}

// SaveGeneratedBills 事务性保存生成的账单：新增、刷新待付账单、跳过已支付或已取消账单。
func (r *RentBillRepository) SaveGeneratedBills(
	ctx context.Context,
	accountID string,
	bills []*domainrentbill.RentBill,
) (*domainrentbill.SaveGeneratedBillsResult, error) {
	if len(bills) == 0 {
		return nil, domainrentbill.ErrBillsRequired
	}
	for _, bill := range bills {
		if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
			return nil, err
		}
	}

	result := &domainrentbill.SaveGeneratedBillsResult{
		Saved:   make([]*domainrentbill.RentBill, 0, len(bills)),
		Skipped: make([]*domainrentbill.RentBill, 0),
	}
	err := query.Use(r.q.UnderlyingDB().WithContext(ctx)).
		Transaction(func(client *query.Query) error {
			for _, bill := range bills {
				record, err := client.RentBill.WithContext(ctx).
					Clauses(clause.Locking{Strength: "UPDATE"}).
					Where(
						client.RentBill.RoomID.Eq(bill.RoomID),
						client.RentBill.PeriodMonth.Eq(bill.PeriodMonth),
						client.RentBill.DeletedAt.IsNull(),
						gen.Columns{
							client.RentBill.RoomID,
						}.In(accountRooms(ctx, client, accountID).Select(client.Room.ID)),
					).
					Take()
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					created, err := r.createRentBillRecord(ctx, client, bill)
					if err != nil {
						if isConstraintError(err) {
							return domainrentbill.ErrAlreadyExists
						}
						return err
					}
					createdBill, err := toDomainRentBill(created)
					if err != nil {
						return err
					}
					result.Saved = append(result.Saved, createdBill)
					continue
				}

				if record.Status != domainrentbill.StatusUnpaid {
					skippedBill, err := toDomainRentBill(record)
					if err != nil {
						return err
					}
					result.Skipped = append(result.Skipped, skippedBill)
					continue
				}
				err = r.updateRentBillRecord(
					client.RentBill.WithContext(ctx).Where(client.RentBill.ID.Eq(record.ID)),
					bill,
				)
				if err != nil {
					if isConstraintError(err) {
						return domainrentbill.ErrAlreadyExists
					}
					return err
				}
				updated, err := client.RentBill.WithContext(ctx).
					Where(client.RentBill.ID.Eq(record.ID)).
					Take()
				if err != nil {
					return err
				}

				updatedBill, err := toDomainRentBill(updated)
				if err != nil {
					return err
				}
				result.Saved = append(result.Saved, updatedBill)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update 更新账号名下指定租金账单。
func (r *RentBillRepository) Update(
	ctx context.Context,
	accountID string,
	bill *domainrentbill.RentBill,
) (*domainrentbill.RentBill, error) {
	if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
		return nil, err
	}
	m := r.q.RentBill
	q := m.WithContext(ctx).Where(m.ID.Eq(bill.ID), m.DeletedAt.IsNull(),
		gen.Columns{m.RoomID}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)))
	if err := r.updateRentBillRecord(q, bill); err != nil {
		return nil, err
	}
	return r.Get(ctx, accountID, bill.ID)
}

// Delete 软删除账号名下指定租金账单。
func (r *RentBillRepository) Delete(ctx context.Context, accountID string, id int) error {
	result, err := r.q.RentBill.WithContext(ctx).
		Where(
			r.q.RentBill.ID.Eq(id),
			r.q.RentBill.DeletedAt.IsNull(),
			gen.Columns{
				r.q.RentBill.RoomID,
			}.In(accountRooms(ctx, r.q, accountID).Select(r.q.Room.ID)),
		).
		Update(r.q.RentBill.DeletedAt, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainrentbill.ErrNotFound
	}
	return nil
}

func (r *RentBillRepository) createRentBillRecord(
	ctx context.Context,
	client *query.Query,
	bill *domainrentbill.RentBill,
) (*models.RentBill, error) {
	record, err := toModelRentBill(bill)
	if err != nil {
		return nil, err
	}
	if err := client.RentBill.WithContext(ctx).Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (r *RentBillRepository) updateRentBillRecord(
	q query.IRentBillDo,
	bill *domainrentbill.RentBill,
) error {
	m := r.q.RentBill
	record, err := toModelRentBill(bill)
	if err != nil {
		return err
	}
	result, err := q.Select(
		m.RoomID,
		m.PeriodMonth,
		m.BaseRent,
		m.MeterFee,
		m.MeterReadingsJSON,
		m.TotalAmount,
		m.Status,
		m.PaidAt,
		m.Remark,
	).Updates(record)
	if err != nil {
		if isConstraintError(err) {
			return domainrentbill.ErrAlreadyExists
		}
		return err
	}
	if result.RowsAffected == 0 {
		count, err := q.Count()
		if err != nil {
			return err
		}
		if count == 0 {
			return domainrentbill.ErrNotFound
		}
	}
	return nil
}

func toModelRentBill(bill *domainrentbill.RentBill) (*models.RentBill, error) {
	var readingsJSON *string
	if bill.MeterReadings != nil {
		data, err := json.Marshal(bill.MeterReadings)
		if err != nil {
			return nil, fmt.Errorf("账单抄表快照编码失败: %w", err)
		}
		encoded := string(data)
		readingsJSON = &encoded
	}
	return &models.RentBill{
		RoomID:            bill.RoomID,
		PeriodMonth:       bill.PeriodMonth,
		BaseRent:          bill.BaseRent,
		MeterFee:          bill.MeterFee,
		MeterReadingsJSON: readingsJSON,
		TotalAmount:       bill.TotalAmount,
		Status:            bill.Status,
		PaidAt:            bill.PaidAt,
		Remark:            bill.Remark,
	}, nil
}

// queryRentBills 构建带账号归属约束的租金账单查询。
func (r *RentBillRepository) queryRentBills(
	ctx context.Context,
	accountID string,
	params domainrentbill.ListQuery,
) query.IRentBillDo {
	rooms := accountRooms(ctx, r.q, accountID)
	if params.UnitID > 0 {
		rooms = rooms.Where(r.q.Room.UnitID.Eq(params.UnitID))
	}
	q := r.q.RentBill.WithContext(ctx).Where(
		r.q.RentBill.DeletedAt.IsNull(),
		gen.Columns{r.q.RentBill.RoomID}.In(rooms.Select(r.q.Room.ID)),
	)

	if params.RoomID > 0 {
		q = q.Where(r.q.RentBill.RoomID.Eq(params.RoomID))
	}
	if params.PeriodMonth != "" {
		q = q.Where(r.q.RentBill.PeriodMonth.Eq(params.PeriodMonth))
	}
	if params.PeriodMonthFrom != "" {
		q = q.Where(r.q.RentBill.PeriodMonth.Gte(params.PeriodMonthFrom))
	}
	if params.PeriodMonthTo != "" {
		q = q.Where(r.q.RentBill.PeriodMonth.Lte(params.PeriodMonthTo))
	}
	if params.Status != "" {
		q = q.Where(r.q.RentBill.Status.Eq(params.Status))
	}
	return q
}

// ensureRentBillRoomInAccount 确认账单房间属于当前账号。
func (r *RentBillRepository) ensureRentBillRoomInAccount(
	ctx context.Context,
	accountID string,
	roomID int,
) error {
	count, err := accountRooms(ctx, r.q, accountID).Where(r.q.Room.ID.Eq(roomID)).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return domainrentbill.ErrRoomNotFound
	}
	return nil
}

// toDomainRentBill 将 GORM 租金账单转换为领域对象。
func toDomainRentBill(record *models.RentBill) (*domainrentbill.RentBill, error) {
	var readings []domainrentbill.MeterReadingSnapshot
	if record.MeterReadingsJSON != nil {
		if err := json.Unmarshal([]byte(*record.MeterReadingsJSON), &readings); err != nil {
			return nil, fmt.Errorf("账单抄表快照解码失败: %w", err)
		}
	}
	return &domainrentbill.RentBill{
		ID:            record.ID,
		RoomID:        record.RoomID,
		PeriodMonth:   record.PeriodMonth,
		BaseRent:      record.BaseRent,
		MeterFee:      record.MeterFee,
		MeterReadings: readings,
		TotalAmount:   record.TotalAmount,
		Status:        record.Status,
		PaidAt:        record.PaidAt,
		Remark:        record.Remark,
		CreatedAt:     record.CreatedAt,
		UpdatedAt:     record.UpdatedAt,
		DeletedAt:     record.DeletedAt,
	}, nil
}

// toDomainRentBills 将 GORM 租金账单列表转换为领域对象列表。
func toDomainRentBills(records []*models.RentBill) ([]*domainrentbill.RentBill, error) {
	bills := make([]*domainrentbill.RentBill, 0, len(records))
	for _, record := range records {
		bill, err := toDomainRentBill(record)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}
