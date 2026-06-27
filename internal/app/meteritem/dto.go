package meteritem

import (
	appcore "git.superwhys.top/rent-manager/rent-manager-server/internal/app"
	domainmeteritem "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/meteritem"
)

// CreateMeterItemCommand 是创建抄表项的请求体。
type CreateMeterItemCommand struct {
	Code      string  `json:"code" validate:"required,max=255"`
	Unit      string  `json:"unit" validate:"required,max=255"`
	UnitPrice float64 `json:"unit_price" validate:"min=0"`
	CalcType  string  `json:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
	Remark    *string `json:"remark"`
}

// UpdateMeterItemCommand 是更新抄表项的请求参数。
type UpdateMeterItemCommand struct {
	MeterItemID int     `uri:"meter_item_id" validate:"required,min=1"`
	Code        string  `json:"code" validate:"required,max=255"`
	Unit        string  `json:"unit" validate:"required,max=255"`
	UnitPrice   float64 `json:"unit_price" validate:"min=0"`
	CalcType    string  `json:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
	Remark      *string `json:"remark"`
}

// ListMeterItemsQuery 是抄表项列表查询参数。
type ListMeterItemsQuery struct {
	appcore.Pagination

	Code     string `form:"code"`
	IsSystem *bool  `form:"is_system"`
	CalcType string `form:"calc_type" validate:"omitempty,oneof=DIFFERENCE DIRECT"`
}

// MeterItemIDQuery 是抄表项 ID 路径参数。
type MeterItemIDQuery struct {
	MeterItemID int `uri:"meter_item_id" validate:"required,min=1"`
}

// MeterItemResponse 是抄表项响应。
type MeterItemResponse struct {
	ID        int     `json:"id"`
	AccountID int     `json:"account_id"`
	Code      string  `json:"code"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unit_price"`
	IsSystem  bool    `json:"is_system"`
	CalcType  string  `json:"calc_type"`
	Remark    *string `json:"remark,omitempty"`
}

// ToMeterItemResponse 将抄表项领域对象转换为响应对象。
func ToMeterItemResponse(item *domainmeteritem.MeterItem) MeterItemResponse {
	if item == nil {
		return MeterItemResponse{}
	}
	return MeterItemResponse{
		ID:        item.ID,
		AccountID: item.AccountID,
		Code:      item.Code,
		Unit:      item.Unit,
		UnitPrice: item.UnitPrice,
		IsSystem:  item.IsSystem,
		CalcType:  item.CalcType,
		Remark:    item.Remark,
	}
}

// ToMeterItemListResponse 将抄表项列表转换为响应列表。
func ToMeterItemListResponse(items []*domainmeteritem.MeterItem) []MeterItemResponse {
	responses := make([]MeterItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, ToMeterItemResponse(item))
	}
	return responses
}
