package response

import (
	"github.com/miebyte/goutils/ginutils"
	domainmeteritem "github.com/superwhys/rent-manager-server/internal/domain/meteritem"
)

// MeterItemResponse 是抄表项响应。
type MeterItemResponse struct {
	ID             int     `json:"id"`
	OrganizationID int     `json:"organization_id"`
	Code           string  `json:"code"`
	Unit           string  `json:"unit"`
	IsSystem       bool    `json:"is_system"`
	CalcType       string  `json:"calc_type"`
	Remark         *string `json:"remark,omitempty"`
}

// ToMeterItemResponse 将抄表项领域对象转换为响应对象。
func ToMeterItemResponse(item *domainmeteritem.MeterItem) MeterItemResponse {
	if item == nil {
		return MeterItemResponse{}
	}
	return MeterItemResponse{
		ID:             item.ID,
		OrganizationID: item.OrganizationID,
		Code:           item.Code,
		Unit:           item.Unit,
		IsSystem:       item.IsSystem,
		CalcType:       item.CalcType,
		Remark:         item.Remark,
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

// MeterItemResponseRet 返回抄表项接口响应。
func MeterItemResponseRet(item *domainmeteritem.MeterItem) *ginutils.Ret[MeterItemResponse] {
	return ResponseWithData(ToMeterItemResponse(item))
}

// MeterItemListResponseRet 返回抄表项列表接口响应。
func MeterItemListResponseRet(items []*domainmeteritem.MeterItem) *ginutils.Ret[[]MeterItemResponse] {
	return ResponseWithData(ToMeterItemListResponse(items))
}
