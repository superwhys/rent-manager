package repo

import (
	"context"

	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
	entclient "github.com/superwhys/rent-manager-server/internal/ent"
	entmeteritem "github.com/superwhys/rent-manager-server/internal/ent/meteritem"
)

type MeterItemRepository struct {
	client *entclient.Client
}

func NewMeterItemRepository(client *entclient.Client) *MeterItemRepository {
	return &MeterItemRepository{client: client}
}

func (r *MeterItemRepository) Create(ctx context.Context, createItem *domainmeteritem.MeterItem) (*domainmeteritem.MeterItem, error) {
	record, err := r.client.MeterItem.Create().
		SetAccountID(createItem.AccountID).
		SetCode(createItem.Code).
		SetUnit(createItem.Unit).
		SetIsSystem(false).
		SetCalcType(entmeteritem.CalcType(createItem.CalcType)).
		SetNillableRemark(createItem.Remark).
		Save(ctx)
	if err != nil {
		return nil, mapMeterItemError(err)
	}

	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) Get(ctx context.Context, accountID int, id int) (*domainmeteritem.MeterItem, error) {
	record, err := r.client.MeterItem.Query().
		Where(entmeteritem.ID(id)).
		Where(entmeteritem.AccountID(accountID)).
		Where(entmeteritem.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, mapMeterItemError(err)
	}

	return toDomainMeterItem(record), nil
}

func (r *MeterItemRepository) List(ctx context.Context, query domainmeteritem.ListQuery) ([]*domainmeteritem.MeterItem, error) {
	builder := r.client.MeterItem.Query().
		Where(entmeteritem.AccountID(query.AccountID)).
		Where(entmeteritem.DeletedAtIsNil()).
		Order(entclient.Desc(entmeteritem.FieldID))

	if query.Code != "" {
		builder = builder.Where(entmeteritem.CodeContainsFold(query.Code))
	}
	if query.IsSystem != nil {
		builder = builder.Where(entmeteritem.IsSystem(*query.IsSystem))
	}
	if query.CalcType != "" {
		builder = builder.Where(entmeteritem.CalcTypeEQ(entmeteritem.CalcType(query.CalcType)))
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
	items := make([]*domainmeteritem.MeterItem, 0, len(records))
	for _, record := range records {
		items = append(items, toDomainMeterItem(record))
	}

	return items, nil
}

func (r *MeterItemRepository) Update(ctx context.Context, updateItem *domainmeteritem.MeterItem) (*domainmeteritem.MeterItem, error) {
	builder := r.client.MeterItem.Update().
		Where(entmeteritem.ID(updateItem.ID)).
		Where(entmeteritem.AccountID(updateItem.AccountID)).
		Where(entmeteritem.DeletedAtIsNil()).
		SetCode(updateItem.Code).
		SetUnit(updateItem.Unit).
		SetCalcType(entmeteritem.CalcType(updateItem.CalcType))
	if updateItem.Remark == nil {
		builder = builder.ClearRemark()
	} else {
		builder = builder.SetRemark(*updateItem.Remark)
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return nil, mapMeterItemError(err)
	}
	if affected == 0 {
		return nil, domainmeteritem.ErrNotFound
	}
	return r.Get(ctx, updateItem.AccountID, updateItem.ID)
}

func ensureMeterItemsOwned(ctx context.Context, client *entclient.Client, accountID int, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	count, err := client.MeterItem.Query().
		Where(entmeteritem.IDIn(ids...)).
		Where(entmeteritem.AccountID(accountID)).
		Where(entmeteritem.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return err
	}
	if count != len(ids) {
		return domainmeteritem.ErrNotFound
	}
	return nil
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
		IsSystem:  record.IsSystem,
		CalcType:  record.CalcType.String(),
		Remark:    record.Remark,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func mapMeterItemError(err error) error {
	if err == nil {
		return nil
	}
	if entclient.IsNotFound(err) {
		return domainmeteritem.ErrNotFound
	}
	if entclient.IsConstraintError(err) {
		return domainmeteritem.ErrCodeAlreadyExists
	}
	return err
}
