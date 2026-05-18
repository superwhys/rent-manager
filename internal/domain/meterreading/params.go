package meterreading

import "time"

type CreateParams struct {
	RoomID          int
	MeterItemID     int
	PeriodMonth     string
	PreviousReading *float64
	CurrentReading  float64
	UsageAmount     *float64
	ReadingTime     time.Time
	ReaderID        *int64
	Source          string
	Remark          *string
}

type UpdateParams struct {
	ID              int
	RoomID          int
	MeterItemID     int
	PeriodMonth     string
	PreviousReading *float64
	CurrentReading  float64
	UsageAmount     *float64
	ReadingTime     time.Time
	ReaderID        *int64
	Source          string
	Remark          *string
}

type ListQuery struct {
	RoomID          int
	MeterItemID     int
	PeriodMonth     string
	PeriodMonthFrom string
	PeriodMonthTo   string
	Source          string
	Limit           int
	Offset          int
}
