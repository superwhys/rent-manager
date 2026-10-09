package meterreading

import "context"

type Repository interface {
	Create(ctx context.Context, accountID string, reading *MeterReading) (*MeterReading, error)
	Get(ctx context.Context, accountID string, id int) (*MeterReading, error)
	GetByRoomMeterItemPeriod(ctx context.Context, accountID string, roomID int, meterItemID int, periodMonth string) (*MeterReading, error)
	GetLatestBeforePeriod(ctx context.Context, accountID string, roomID int, meterItemID int, periodMonth string) (*MeterReading, error)
	List(ctx context.Context, accountID string, query ListQuery) ([]*MeterReading, error)
	SaveRoomMonth(ctx context.Context, accountID string, readings []*MeterReading) ([]*MeterReading, error)
	Update(ctx context.Context, accountID string, reading *MeterReading) (*MeterReading, error)
	Delete(ctx context.Context, accountID string, id int) error
}
