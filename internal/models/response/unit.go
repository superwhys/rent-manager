package response

import (
	"github.com/miebyte/goutils/ginutils"
	domainunit "github.com/superwhys/rent-manager-server/internal/domain/unit"
)

// UnitMeterItemResponse 是单元默认抄表项响应。
type UnitMeterItemResponse struct {
	ID       int     `json:"id"`
	Code     string  `json:"code"`
	Unit     string  `json:"unit"`
	IsSystem bool    `json:"is_system"`
	CalcType string  `json:"calc_type"`
	Remark   *string `json:"remark,omitempty"`
}

// UnitResponse 是单元响应。
type UnitResponse struct {
	ID                int                     `json:"id"`
	AccountID         int                     `json:"account_id"`
	Name              string                  `json:"name"`
	Address           string                  `json:"address"`
	City              string                  `json:"city"`
	Province          string                  `json:"province"`
	DefaultMeterItems []UnitMeterItemResponse `json:"default_meter_items"`
}

// ToUnitResponse 将单元领域对象转换为响应对象。
func ToUnitResponse(unit *domainunit.Unit) UnitResponse {
	if unit == nil {
		return UnitResponse{}
	}
	return UnitResponse{
		ID:                unit.ID,
		AccountID:         unit.AccountID,
		Name:              unit.Name,
		Address:           unit.Address,
		City:              unit.City,
		Province:          unit.Province,
		DefaultMeterItems: toUnitMeterItemResponses(unit.DefaultMeterItems),
	}
}

// ToUnitListResponse 将单元列表转换为响应列表。
func ToUnitListResponse(units []*domainunit.Unit) []UnitResponse {
	ret := make([]UnitResponse, 0, len(units))
	for _, unit := range units {
		ret = append(ret, ToUnitResponse(unit))
	}
	return ret
}

// UnitResponseRet 返回单元接口响应。
func UnitResponseRet(unit *domainunit.Unit) *ginutils.Ret[UnitResponse] {
	return ResponseWithData(ToUnitResponse(unit))
}

// UnitListResponseRet 返回单元列表接口响应。
func UnitListResponseRet(units []*domainunit.Unit) *ginutils.Ret[[]UnitResponse] {
	return ResponseWithData(ToUnitListResponse(units))
}

func toUnitMeterItemResponses(items []*domainunit.MeterItem) []UnitMeterItemResponse {
	ret := make([]UnitMeterItemResponse, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ret = append(ret, UnitMeterItemResponse{
			ID:       item.ID,
			Code:     item.Code,
			Unit:     item.Unit,
			IsSystem: item.IsSystem,
			CalcType: item.CalcType,
			Remark:   item.Remark,
		})
	}
	return ret
}
