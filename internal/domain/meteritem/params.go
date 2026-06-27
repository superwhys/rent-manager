package meteritem

type CreateParams struct {
	AccountID int
	Code      string
	Unit      string
	UnitPrice float64
	CalcType  string
	Remark    *string
}

type UpdateParams struct {
	ID        int
	AccountID int
	Code      string
	Unit      string
	UnitPrice float64
	CalcType  string
	Remark    *string
}

type ListQuery struct {
	AccountID int
	Code      string
	IsSystem  *bool
	CalcType  string
	Limit     int
	Offset    int
}
