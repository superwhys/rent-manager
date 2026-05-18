package rentbill

import "time"

type CreateParams struct {
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	Status      string
	PaidAt      *time.Time
	Remark      *string
}

type GenerateParams struct {
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	Remark      *string
}

type UpdateParams struct {
	ID          int
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	Status      string
	PaidAt      *time.Time
	Remark      *string
}

type ListQuery struct {
	RoomID          int
	PeriodMonth     string
	PeriodMonthFrom string
	PeriodMonthTo   string
	Status          string
	Limit           int
	Offset          int
}
