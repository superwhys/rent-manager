package meterreading

import "context"

type Repository interface {
	Create(ctx context.Context, accountID int, reading *MeterReading) (*MeterReading, error)
	Get(ctx context.Context, accountID int, id int) (*MeterReading, error)
	GetByRoomMeterItemPeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*MeterReading, error)
	GetLatestBeforePeriod(ctx context.Context, accountID int, roomID int, meterItemID int, periodMonth string) (*MeterReading, error)
	List(ctx context.Context, accountID int, query ListQuery) ([]*MeterReading, error)
	SaveRoomMonth(ctx context.Context, accountID int, readings []*MeterReading) ([]*MeterReading, error)
	Update(ctx context.Context, accountID int, reading *MeterReading) (*MeterReading, error)
	Delete(ctx context.Context, accountID int, id int) error
}
