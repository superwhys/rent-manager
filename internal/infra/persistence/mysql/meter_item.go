package mysql

import (
	"context"
	"errors"

	"gorm.io/gen/field"
	"gorm.io/gorm"

	domainmeteritem "github.com/superwhys/rent-manager/internal/domain/meteritem"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type MeterItemRepository struct {
	q *query.Query
}

func NewMeterItemRepository(client *gorm.DB) *MeterItemRepository {
	return &MeterItemRepository{q: query.Use(client)}
}

func (r *MeterItemRepository) Create(
	ctx context.Context,
	item *domainmeteritem.MeterItem,
) (*domainmeteritem.MeterItem, error) {
	record := &models.MeterItem{
		AccountID: item.AccountID,
		Code:      item.Code,
		Unit:      item.Unit,
		UnitPrice: item.UnitPrice,
		CalcType:  item.CalcType,
		Remark:    item.Remark,
		IsSystem:  item.IsSystem,
	}
	err := r.q.MeterItem.WithContext(ctx).Create(record)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainmeteritem.ErrCodeAlreadyExists
		}
		return nil, err
	}
	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) Get(
	ctx context.Context,
	accountID string,
	id int,
) (*domainmeteritem.MeterItem, error) {
	record, err := r.q.MeterItem.WithContext(ctx).
		Where(r.q.MeterItem.ID.Eq(id), r.q.MeterItem.AccountID.Eq(accountID), r.q.MeterItem.DeletedAt.IsNull()).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainmeteritem.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) List(
	ctx context.Context,
	query domainmeteritem.ListQuery,
) ([]*domainmeteritem.MeterItem, error) {
	q := r.q.MeterItem.WithContext(ctx).
		Where(r.q.MeterItem.AccountID.Eq(query.AccountID), r.q.MeterItem.DeletedAt.IsNull())
	if query.Code != "" {
		q = q.Where(contains(r.q.MeterItem.Code, query.Code))
	}
	if query.IsSystem != nil {
		q = q.Where(r.q.MeterItem.IsSystem.Is(*query.IsSystem))
	}
	if query.CalcType != "" {
		q = q.Where(r.q.MeterItem.CalcType.Eq(query.CalcType))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(r.q.MeterItem.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainMeterItems(records), nil
}

func (r *MeterItemRepository) Update(
	ctx context.Context,
	item *domainmeteritem.MeterItem,
) (*domainmeteritem.MeterItem, error) {
	m := r.q.MeterItem
	assignments := []field.AssignExpr{
		m.Code.Value(item.Code),
		m.Unit.Value(item.Unit),
		m.UnitPrice.Value(item.UnitPrice),
		m.CalcType.Value(item.CalcType),
	}
	if item.Remark != nil {
		assignments = append(assignments, m.Remark.Value(*item.Remark))
	}
	result, err := m.WithContext(ctx).
		Where(m.ID.Eq(item.ID), m.AccountID.Eq(item.AccountID), m.DeletedAt.IsNull()).
		UpdateSimple(assignments...)
	if err != nil {
		if isConstraintError(err) {
			return nil, domainmeteritem.ErrCodeAlreadyExists
		}
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, domainmeteritem.ErrNotFound
	}
	return r.Get(ctx, item.AccountID, item.ID)
}

func toDomainMeterItem(record *models.MeterItem) *domainmeteritem.MeterItem {
	return &domainmeteritem.MeterItem{
		ID:        record.ID,
		AccountID: record.AccountID,
		Code:      record.Code,
		Unit:      record.Unit,
		UnitPrice: record.UnitPrice,
		IsSystem:  record.IsSystem,
		CalcType:  record.CalcType,
		Remark:    record.Remark,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func toDomainMeterItems(records []*models.MeterItem) []*domainmeteritem.MeterItem {
	items := make([]*domainmeteritem.MeterItem, 0, len(records))
	for _, record := range records {
		items = append(items, toDomainMeterItem(record))
	}
	return items
}
