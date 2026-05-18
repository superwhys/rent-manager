package meteritem

// CreateParams 描述创建抄表项时需要满足的领域输入。
type CreateParams struct {
	AccountID int
	Code      string
	Unit      string
	UnitPrice int
	CalcType  string
	Remark    *string
}

// UpdateParams 描述更新抄表项时允许变更的领域输入。
type UpdateParams struct {
	ID        int
	AccountID int
	Code      string
	Unit      string
	UnitPrice int
	CalcType  string
	Remark    *string
}

// ListQuery 描述抄表项列表查询条件。
type ListQuery struct {
	AccountID int
	Code      string
	IsSystem  *bool
	CalcType  string
	Limit     int
	Offset    int
}
