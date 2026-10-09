package mysql

import (
	"context"
	"time"

	"github.com/superwhys/rent-manager/internal/infra/persistence/ent/predicate"

	domainrentbill "github.com/superwhys/rent-manager/internal/domain/rentbill"
	entclient "github.com/superwhys/rent-manager/internal/infra/persistence/ent"
	entrentbill "github.com/superwhys/rent-manager/internal/infra/persistence/ent/rentbill"
	entroom "github.com/superwhys/rent-manager/internal/infra/persistence/ent/room"
	entunit "github.com/superwhys/rent-manager/internal/infra/persistence/ent/unit"
)

// RentBillRepository 提供租金账单的 Ent 持久化实现。
type RentBillRepository struct {
	client *entclient.Client
}

// NewRentBillRepository 创建租金账单仓储。
func NewRentBillRepository(client *entclient.Client) *RentBillRepository {
	return &RentBillRepository{client: client}
}

// Create 创建账号名下房间的租金账单。
func (r *RentBillRepository) Create(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
		return nil, err
	}
	record, err := r.client.RentBill.Create().
		SetRoomID(bill.RoomID).
		SetPeriodMonth(bill.PeriodMonth).
		SetBaseRent(bill.BaseRent).
		SetMeterFee(bill.MeterFee).
		SetTotalAmount(bill.TotalAmount).
		SetStatus(entrentbill.Status(bill.Status)).
		SetNillablePaidAt(bill.PaidAt).
		SetNillableRemark(bill.Remark).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainrentbill.ErrAlreadyExists
		}
		return nil, err
	}
	return r.Get(ctx, accountID, record.ID)
}

// Get 获取账号名下指定租金账单。
func (r *RentBillRepository) Get(ctx context.Context, accountID int, id int) (*domainrentbill.RentBill, error) {
	record, err := r.client.RentBill.Query().
		Where(
			entrentbill.ID(id),
			entrentbill.DeletedAtIsNil(),
			entrentbill.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record), nil
}

// GetByID 获取指定租金账单，不做账号归属过滤。
func (r *RentBillRepository) GetByID(ctx context.Context, id int) (*domainrentbill.RentBill, error) {
	record, err := r.client.RentBill.Query().
		Where(entrentbill.ID(id), entrentbill.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record), nil
}

// GetByRoomPeriod 获取账号名下房间指定月份的租金账单。
func (r *RentBillRepository) GetByRoomPeriod(ctx context.Context, accountID int, roomID int, periodMonth string) (*domainrentbill.RentBill, error) {
	record, err := r.client.RentBill.Query().
		Where(
			entrentbill.RoomID(roomID),
			entrentbill.PeriodMonth(periodMonth),
			entrentbill.DeletedAtIsNil(),
			entrentbill.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainrentbill.ErrNotFound
		}
		return nil, err
	}
	return toDomainRentBill(record), nil
}

// List 查询账号名下的租金账单列表。
func (r *RentBillRepository) List(ctx context.Context, accountID int, query domainrentbill.ListQuery) ([]*domainrentbill.RentBill, error) {
	q := r.queryRentBills(accountID, query)
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entrentbill.FieldPeriodMonth), entclient.Desc(entrentbill.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRentBills(records), nil
}

// SaveGeneratedBills 事务性保存生成的账单：新增、刷新待付账单、跳过已支付或已取消账单。
func (r *RentBillRepository) SaveGeneratedBills(ctx context.Context, accountID int, bills []*domainrentbill.RentBill) (*domainrentbill.SaveGeneratedBillsResult, error) {
	if len(bills) == 0 {
		return nil, domainrentbill.ErrBillsRequired
	}
	for _, bill := range bills {
		if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
			return nil, err
		}
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
	result := &domainrentbill.SaveGeneratedBillsResult{
		Saved:   make([]*domainrentbill.RentBill, 0, len(bills)),
		Skipped: make([]*domainrentbill.RentBill, 0),
	}
	for _, bill := range bills {
		record, err := client.RentBill.Query().
			Where(
				entrentbill.RoomID(bill.RoomID),
				entrentbill.PeriodMonth(bill.PeriodMonth),
				entrentbill.DeletedAtIsNil(),
				entrentbill.HasRoomWith(r.accountRoomPredicates(accountID)...),
			).
			Only(ctx)
		if err != nil {
			if !entclient.IsNotFound(err) {
				return nil, err
			}
			created, err := r.createRentBillRecord(ctx, client, bill)
			if err != nil {
				if entclient.IsConstraintError(err) {
					return nil, domainrentbill.ErrAlreadyExists
				}
				return nil, err
			}
			result.Saved = append(result.Saved, toDomainRentBill(created))
			continue
		}

		if record.Status.String() != domainrentbill.StatusUnpaid {
			result.Skipped = append(result.Skipped, toDomainRentBill(record))
			continue
		}
		updated, err := r.updateRentBillRecord(ctx, client, record.ID, bill)
		if err != nil {
			if entclient.IsConstraintError(err) {
				return nil, domainrentbill.ErrAlreadyExists
			}
			return nil, err
		}
		result.Saved = append(result.Saved, toDomainRentBill(updated))
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return result, nil
}

// Update 更新账号名下指定租金账单。
func (r *RentBillRepository) Update(ctx context.Context, accountID int, bill *domainrentbill.RentBill) (*domainrentbill.RentBill, error) {
	if err := r.ensureRentBillRoomInAccount(ctx, accountID, bill.RoomID); err != nil {
		return nil, err
	}
	updater := r.client.RentBill.Update().
		Where(
			entrentbill.ID(bill.ID),
			entrentbill.DeletedAtIsNil(),
			entrentbill.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		SetRoomID(bill.RoomID).
		SetPeriodMonth(bill.PeriodMonth).
		SetBaseRent(bill.BaseRent).
		SetMeterFee(bill.MeterFee).
		SetTotalAmount(bill.TotalAmount).
		SetStatus(entrentbill.Status(bill.Status))
	if bill.PaidAt == nil {
		updater = updater.ClearPaidAt()
	} else {
		updater = updater.SetPaidAt(*bill.PaidAt)
	}
	if bill.Remark == nil {
		updater = updater.ClearRemark()
	} else {
		updater = updater.SetRemark(*bill.Remark)
	}
	affected, err := updater.Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainrentbill.ErrAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainrentbill.ErrNotFound
	}
	return r.Get(ctx, accountID, bill.ID)
}

// Delete 软删除账号名下指定租金账单。
func (r *RentBillRepository) Delete(ctx context.Context, accountID int, id int) error {
	affected, err := r.client.RentBill.Update().
		Where(
			entrentbill.ID(id),
			entrentbill.DeletedAtIsNil(),
			entrentbill.HasRoomWith(r.accountRoomPredicates(accountID)...),
		).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domainrentbill.ErrNotFound
	}
	return nil
}

func (r *RentBillRepository) createRentBillRecord(ctx context.Context, client *entclient.Client, bill *domainrentbill.RentBill) (*entclient.RentBill, error) {
	return client.RentBill.Create().
		SetRoomID(bill.RoomID).
		SetPeriodMonth(bill.PeriodMonth).
		SetBaseRent(bill.BaseRent).
		SetMeterFee(bill.MeterFee).
		SetTotalAmount(bill.TotalAmount).
		SetStatus(entrentbill.Status(bill.Status)).
		SetNillablePaidAt(bill.PaidAt).
		SetNillableRemark(bill.Remark).
		Save(ctx)
}

func (r *RentBillRepository) updateRentBillRecord(ctx context.Context, client *entclient.Client, id int, bill *domainrentbill.RentBill) (*entclient.RentBill, error) {
	updater := client.RentBill.UpdateOneID(id).
		SetRoomID(bill.RoomID).
		SetPeriodMonth(bill.PeriodMonth).
		SetBaseRent(bill.BaseRent).
		SetMeterFee(bill.MeterFee).
		SetTotalAmount(bill.TotalAmount).
		SetStatus(entrentbill.Status(bill.Status))
	if bill.PaidAt == nil {
		updater = updater.ClearPaidAt()
	} else {
		updater = updater.SetPaidAt(*bill.PaidAt)
	}
	if bill.Remark == nil {
		updater = updater.ClearRemark()
	} else {
		updater = updater.SetRemark(*bill.Remark)
	}
	return updater.Save(ctx)
}

// queryRentBills 构建带账号归属约束的租金账单查询。
func (r *RentBillRepository) queryRentBills(accountID int, query domainrentbill.ListQuery) *entclient.RentBillQuery {
	roomPredicates := r.accountRoomPredicates(accountID)
	if query.UnitID > 0 {
		roomPredicates = append(roomPredicates, entroom.UnitID(query.UnitID))
	}
	q := r.client.RentBill.Query().
		Where(
			entrentbill.DeletedAtIsNil(),
			entrentbill.HasRoomWith(roomPredicates...),
		)
	if query.RoomID > 0 {
		q = q.Where(entrentbill.RoomID(query.RoomID))
	}
	if query.PeriodMonth != "" {
		q = q.Where(entrentbill.PeriodMonth(query.PeriodMonth))
	}
	if query.PeriodMonthFrom != "" {
		q = q.Where(entrentbill.PeriodMonthGTE(query.PeriodMonthFrom))
	}
	if query.PeriodMonthTo != "" {
		q = q.Where(entrentbill.PeriodMonthLTE(query.PeriodMonthTo))
	}
	if query.Status != "" {
		q = q.Where(entrentbill.StatusEQ(entrentbill.Status(query.Status)))
	}
	return q
}

// accountRoomPredicates 返回账号名下未删除房间的查询条件。
func (r *RentBillRepository) accountRoomPredicates(accountID int) []predicate.Room {
	return []predicate.Room{
		entroom.DeletedAtIsNil(),
		entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
	}
}

// ensureRentBillRoomInAccount 确认账单房间属于当前账号。
func (r *RentBillRepository) ensureRentBillRoomInAccount(ctx context.Context, accountID int, roomID int) error {
	exists, err := r.client.Room.Query().
		Where(
			entroom.ID(roomID),
			entroom.DeletedAtIsNil(),
			entroom.HasUnitWith(entunit.AccountID(accountID), entunit.DeletedAtIsNil()),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domainrentbill.ErrRoomNotFound
	}
	return nil
}

// toDomainRentBill 将 Ent 租金账单转换为领域对象。
func toDomainRentBill(record *entclient.RentBill) *domainrentbill.RentBill {
	if record == nil {
		return nil
	}
	return &domainrentbill.RentBill{
		ID:          record.ID,
		RoomID:      record.RoomID,
		PeriodMonth: record.PeriodMonth,
		BaseRent:    record.BaseRent,
		MeterFee:    record.MeterFee,
		TotalAmount: record.TotalAmount,
		Status:      record.Status.String(),
		PaidAt:      record.PaidAt,
		Remark:      record.Remark,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		DeletedAt:   record.DeletedAt,
	}
}

// toDomainRentBills 将 Ent 租金账单列表转换为领域对象列表。
func toDomainRentBills(records []*entclient.RentBill) []*domainrentbill.RentBill {
	bills := make([]*domainrentbill.RentBill, 0, len(records))
	for _, record := range records {
		bills = append(bills, toDomainRentBill(record))
	}
	return bills
}
