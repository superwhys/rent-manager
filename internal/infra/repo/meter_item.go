package repo

import (
	"context"

	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
	entclient "git.superwhys.top/rent-manager/rent-manager-server/internal/ent"
	entmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/ent/meteritem"
)

type MeterItemRepository struct {
	client *entclient.Client
}

func NewMeterItemRepository(client *entclient.Client) *MeterItemRepository {
	return &MeterItemRepository{client: client}
}

func (r *MeterItemRepository) Create(ctx context.Context, item *domainmeteritem.MeterItem) (*domainmeteritem.MeterItem, error) {
	record, err := r.client.MeterItem.Create().
		SetAccountID(item.AccountID).
		SetCode(item.Code).
		SetUnit(item.Unit).
		SetUnitPrice(item.UnitPrice).
		SetCalcType(entmeteritem.CalcType(item.CalcType)).
		SetNillableRemark(item.Remark).
		SetIsSystem(item.IsSystem).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainmeteritem.ErrCodeAlreadyExists
		}
		return nil, err
	}
	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) Get(ctx context.Context, accountID int, id int) (*domainmeteritem.MeterItem, error) {
	record, err := r.client.MeterItem.Query().
		Where(entmeteritem.ID(id), entmeteritem.AccountID(accountID), entmeteritem.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if entclient.IsNotFound(err) {
			return nil, domainmeteritem.ErrNotFound
		}
		return nil, err
	}
	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) List(ctx context.Context, query domainmeteritem.ListQuery) ([]*domainmeteritem.MeterItem, error) {
	q := r.client.MeterItem.Query().
		Where(entmeteritem.AccountID(query.AccountID), entmeteritem.DeletedAtIsNil())
	if query.Code != "" {
		q = q.Where(entmeteritem.CodeContains(query.Code))
	}
	if query.IsSystem != nil {
		q = q.Where(entmeteritem.IsSystem(*query.IsSystem))
	}
	if query.CalcType != "" {
		q = q.Where(entmeteritem.CalcTypeEQ(entmeteritem.CalcType(query.CalcType)))
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	records, err := q.Order(entclient.Desc(entmeteritem.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainMeterItems(records), nil
}

func (r *MeterItemRepository) Update(ctx context.Context, item *domainmeteritem.MeterItem) (*domainmeteritem.MeterItem, error) {
	affected, err := r.client.MeterItem.Update().
		Where(entmeteritem.ID(item.ID), entmeteritem.AccountID(item.AccountID), entmeteritem.DeletedAtIsNil()).
		SetCode(item.Code).
		SetUnit(item.Unit).
		SetUnitPrice(item.UnitPrice).
		SetCalcType(entmeteritem.CalcType(item.CalcType)).
		SetNillableRemark(item.Remark).
		Save(ctx)
	if err != nil {
		if entclient.IsConstraintError(err) {
			return nil, domainmeteritem.ErrCodeAlreadyExists
		}
		return nil, err
	}
	if affected == 0 {
		return nil, domainmeteritem.ErrNotFound
	}
	return r.Get(ctx, item.AccountID, item.ID)
}

func toDomainMeterItem(record *entclient.MeterItem) *domainmeteritem.MeterItem {
	if record == nil {
		return nil
	}
	return &domainmeteritem.MeterItem{
		ID:        record.ID,
		AccountID: record.AccountID,
		Code:      record.Code,
		Unit:      record.Unit,
		UnitPrice: record.UnitPrice,
		IsSystem:  record.IsSystem,
		CalcType:  record.CalcType.String(),
		Remark:    record.Remark,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func toDomainMeterItems(records []*entclient.MeterItem) []*domainmeteritem.MeterItem {
	items := make([]*domainmeteritem.MeterItem, 0, len(records))
	for _, record := range records {
		items = append(items, toDomainMeterItem(record))
	}
	return items
}
