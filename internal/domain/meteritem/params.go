package meteritem

type CreateParams struct {
	AccountID string
	Code      string
	Unit      string
	UnitPrice float64
	CalcType  string
	Remark    *string
}

type UpdateParams struct {
	ID        int
	AccountID string
	Code      string
	Unit      string
	UnitPrice float64
	CalcType  string
	Remark    *string
}

type ListQuery struct {
	AccountID string
	Code      string
	IsSystem  *bool
	CalcType  string
	Limit     int
	Offset    int
}
