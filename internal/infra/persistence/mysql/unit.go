package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domainunit "github.com/superwhys/rent-manager/internal/domain/unit"
	"github.com/superwhys/rent-manager/internal/infra/persistence/models"
	"github.com/superwhys/rent-manager/internal/infra/persistence/mysql/query"
)

type UnitRepository struct {
	q *query.Query
}

func NewUnitRepository(db *gorm.DB) *UnitRepository {
	return &UnitRepository{q: query.Use(db)}
}

func (r *UnitRepository) Create(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	if err := r.ensureMeterItemsInAccount(ctx, unit.AccountID, unit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}
	jsonExtend, err := unit.Config.JSONString()
	if err != nil {
		return nil, err
	}
	if err := validateByteLength(255, unit.Name, unit.Address, unit.City, unit.Province); err != nil {
		return nil, err
	}
	if err := validateByteLength(2048, jsonExtend); err != nil {
		return nil, err
	}
	record := &models.Unit{
		AccountID:  unit.AccountID,
		Name:       unit.Name,
		Address:    unit.Address,
		City:       unit.City,
		Province:   unit.Province,
		JSONExtend: jsonExtend,
	}
	err = query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		if err := tx.Unit.WithContext(ctx).Create(record); err != nil {
			return err
		}
		return replaceUnitMeterItems(ctx, tx, record.ID, unit.DefaultMeterItemIDs)
	})
	if err != nil {
		if isConstraintError(err) {
			return nil, domainunit.ErrNameAlreadyExists
		}
		return nil, err
	}
	return r.Get(ctx, unit.AccountID, record.ID)
}

func (r *UnitRepository) Get(ctx context.Context, accountID string, id int) (*domainunit.Unit, error) {
	u := r.q.Unit
	record, err := u.WithContext(ctx).
		Where(u.ID.Eq(id), u.AccountID.Eq(accountID), u.DeletedAt.IsNull()).
		Preload(u.MeterItems).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainunit.ErrNotFound
		}
		return nil, err
	}
	return toDomainUnitForAccount(record, accountID)
}

func (r *UnitRepository) GetByID(ctx context.Context, id int) (*domainunit.Unit, error) {
	u := r.q.Unit
	record, err := u.WithContext(ctx).
		Where(u.ID.Eq(id), u.DeletedAt.IsNull()).
		Preload(u.MeterItems).
		Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainunit.ErrNotFound
		}
		return nil, err
	}
	return toDomainUnit(record)
}

func (r *UnitRepository) List(ctx context.Context, params domainunit.ListQuery) ([]*domainunit.Unit, error) {
	u := r.q.Unit
	q := u.WithContext(ctx).
		Where(u.DeletedAt.IsNull()).
		Preload(u.MeterItems)
	switch params.Scope {
	case domainunit.ListScopeOwned:
		q = q.Where(u.AccountID.Eq(params.AccountID))
	case domainunit.ListScopeGranted:
		q = q.Where(r.grantedUnitPredicate(ctx, params.AccountID))
	case domainunit.ListScopeAll, "":
		q = q.Where(u.WithContext(ctx).Where(u.AccountID.Eq(params.AccountID)).Or(r.grantedUnitPredicate(ctx, params.AccountID)))
	default:
		return nil, domainunit.ErrInvalidListScope
	}
	records, err := q.Order(u.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	return toDomainUnitsForAccount(records, params.AccountID)
}

func (r *UnitRepository) grantedUnitPredicate(ctx context.Context, accountID string) gen.Condition {
	grant := r.q.UnitAccountRole
	role := r.q.Role
	return gen.Exists(grant.WithContext(ctx).Where(
		grant.UnitID.EqCol(r.q.Unit.ID),
		grant.AccountID.Eq(accountID),
		grant.Status.Eq("active"),
		grant.DeletedAt.IsNull(),
		gen.Exists(role.WithContext(ctx).Where(
			role.ID.EqCol(grant.RoleID),
			role.IsActive.Is(true),
			role.DeletedAt.IsNull(),
		)),
	))
}

func (r *UnitRepository) Update(ctx context.Context, unit *domainunit.Unit) (*domainunit.Unit, error) {
	if err := r.ensureMeterItemsInAccount(ctx, unit.AccountID, unit.DefaultMeterItemIDs); err != nil {
		return nil, err
	}
	jsonExtend, err := unit.Config.JSONString()
	if err != nil {
		return nil, err
	}
	if err := validateByteLength(255, unit.Name, unit.Address, unit.City, unit.Province); err != nil {
		return nil, err
	}
	if err := validateByteLength(2048, jsonExtend); err != nil {
		return nil, err
	}
	err = query.Use(r.q.UnderlyingDB().WithContext(ctx)).Transaction(func(tx *query.Query) error {
		u := tx.Unit
		target := u.WithContext(ctx).
			Where(u.ID.Eq(unit.ID), u.AccountID.Eq(unit.AccountID), u.DeletedAt.IsNull())
		if _, err := target.Select(u.ID).Clauses(clause.Locking{Strength: "UPDATE"}).Take(); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainunit.ErrNotFound
			}
			return err
		}
		_, err := u.WithContext(ctx).
			Where(u.ID.Eq(unit.ID), u.AccountID.Eq(unit.AccountID), u.DeletedAt.IsNull()).
			UpdateSimple(
				u.Name.Value(unit.Name),
				u.Address.Value(unit.Address),
				u.City.Value(unit.City),
				u.Province.Value(unit.Province),
				u.JSONExtend.Value(jsonExtend),
			)
		if err != nil {
			return err
		}
		return replaceUnitMeterItems(ctx, tx, unit.ID, unit.DefaultMeterItemIDs)
	})
	if err != nil {
		if isConstraintError(err) {
			return nil, domainunit.ErrNameAlreadyExists
		}
		return nil, err
	}
	return r.Get(ctx, unit.AccountID, unit.ID)
}

func (r *UnitRepository) Delete(ctx context.Context, accountID string, id int) error {
	room := r.q.Room
	count, err := room.WithContext(ctx).
		Where(room.UnitID.Eq(id), room.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return domainunit.ErrUnitHasRooms
	}
	u := r.q.Unit
	result, err := u.WithContext(ctx).
		Where(u.ID.Eq(id), u.AccountID.Eq(accountID), u.DeletedAt.IsNull()).
		UpdateSimple(u.DeletedAt.Value(time.Now()))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return domainunit.ErrNotFound
	}
	return nil
}

func (r *UnitRepository) ensureMeterItemsInAccount(ctx context.Context, accountID string, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	item := r.q.MeterItem
	count, err := item.WithContext(ctx).
		Where(item.ID.In(ids...), item.AccountID.Eq(accountID), item.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return domainunit.ErrMeterItemNotFound
	}
	return nil
}

func replaceUnitMeterItems(ctx context.Context, q *query.Query, unitID int, ids []int) error {
	item := q.UnitMeterItem
	if _, err := item.WithContext(ctx).Where(item.UnitID.Eq(unitID)).Delete(); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	records := make([]*models.UnitMeterItem, 0, len(ids))
	for _, id := range ids {
		records = append(records, &models.UnitMeterItem{UnitID: unitID, MeterItemID: id})
	}
	return item.WithContext(ctx).Create(records...)
}

func toDomainUnit(record *models.Unit) (*domainunit.Unit, error) {
	if record == nil {
		return nil, nil
	}
	config, err := domainunit.ParseUnitConfigJSON(record.JSONExtend)
	if err != nil {
		return nil, err
	}
	unit := &domainunit.Unit{
		ID:        record.ID,
		AccountID: record.AccountID,
		Name:      record.Name,
		Address:   record.Address,
		City:      record.City,
		Province:  record.Province,
		Config:    config,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
	if record.MeterItems != nil {
		unit.DefaultMeterItems = make([]*domainunit.MeterItem, 0, len(record.MeterItems))
		unit.DefaultMeterItemIDs = make([]int, 0, len(record.MeterItems))
		for _, item := range record.MeterItems {
			unit.DefaultMeterItemIDs = append(unit.DefaultMeterItemIDs, item.ID)
			unit.DefaultMeterItems = append(unit.DefaultMeterItems, &domainunit.MeterItem{
				ID:        item.ID,
				Code:      item.Code,
				Unit:      item.Unit,
				UnitPrice: item.UnitPrice,
				IsSystem:  item.IsSystem,
				CalcType:  item.CalcType,
				Remark:    item.Remark,
			})
		}
	}
	return unit, nil
}

func toDomainUnitForAccount(record *models.Unit, accountID string) (*domainunit.Unit, error) {
	unit, err := toDomainUnit(record)
	if err != nil {
		return nil, err
	}
	if unit == nil {
		return nil, nil
	}
	if unit.AccountID == accountID {
		unit.AccessType = domainunit.AccessTypeOwned
	} else {
		unit.AccessType = domainunit.AccessTypeGranted
	}
	return unit, nil
}

func toDomainUnitsForAccount(records []*models.Unit, accountID string) ([]*domainunit.Unit, error) {
	units := make([]*domainunit.Unit, 0, len(records))
	for _, record := range records {
		unit, err := toDomainUnitForAccount(record, accountID)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, nil
}
