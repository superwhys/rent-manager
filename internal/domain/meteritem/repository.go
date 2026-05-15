package meteritem

import "context"

type Repository interface {
	Create(ctx context.Context, item *MeterItem) (*MeterItem, error)
	Get(ctx context.Context, accountID int, id int) (*MeterItem, error)
	List(ctx context.Context, query ListQuery) ([]*MeterItem, error)
	Update(ctx context.Context, item *MeterItem) (*MeterItem, error)
}
