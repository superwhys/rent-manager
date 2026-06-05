package rentbill

import "time"

// CreateParams 描述创建租金账单时需要满足的领域输入。
type CreateParams struct {
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	Status      string
	PaidAt      *time.Time
	Remark      *string
}

// GenerateParams 描述按房间与月份生成租金账单时需要的输入。
type GenerateParams struct {
	RoomID      int
	PeriodMonth string
	BaseRent    int
	MeterFee    int
	Remark      *string
}

// UpdateParams 描述更新租金账单时允许变更的领域输入。
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

// ListQuery 描述租金账单列表查询条件。
type ListQuery struct {
	UnitID          int
	RoomID          int
	PeriodMonth     string
	PeriodMonthFrom string
	PeriodMonthTo   string
	Status          string
	Limit           int
	Offset          int
}
