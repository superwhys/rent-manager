package requests

// CreateMeterItemCommand 是创建抄表项的请求体。
type CreateMeterItemCommand struct {
	OrganizationID int     `json:"organization_id" validate:"required,min=1"`
	Code           string  `json:"code" validate:"required,max=255"`
	Unit           string  `json:"unit" validate:"required,max=255"`
	CalcType       string  `json:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
	Remark         *string `json:"remark"`
}

// UpdateMeterItemCommand 是更新抄表项的请求参数。
type UpdateMeterItemCommand struct {
	OrganizationID int     `json:"organization_id" validate:"required,min=1"`
	MeterItemID    int     `uri:"meter_item_id" validate:"required,min=1"`
	Code           string  `json:"code" validate:"required,max=255"`
	Unit           string  `json:"unit" validate:"required,max=255"`
	CalcType       string  `json:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
	Remark         *string `json:"remark"`
}

// ListMeterItemsQuery 是抄表项列表查询参数。
type ListMeterItemsQuery struct {
	Pagination

	OrganizationID int    `form:"organization_id" validate:"required,min=1"`
	Code           string `form:"code"`
	IsSystem       *bool  `form:"is_system"`
	CalcType       string `form:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
}
