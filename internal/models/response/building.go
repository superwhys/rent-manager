package response

import (
	"github.com/miebyte/goutils/ginutils"
	domainbuilding "github.com/superwhys/rent-manager-server/internal/domain/building"
)

// BuildingMeterItemResponse 是楼栋默认抄表项响应。
type BuildingMeterItemResponse struct {
	ID       int     `json:"id"`
	Code     string  `json:"code"`
	Unit     string  `json:"unit"`
	IsSystem bool    `json:"is_system"`
	CalcType string  `json:"calc_type"`
	Remark   *string `json:"remark,omitempty"`
}

// BuildingResponse 是楼栋响应。
type BuildingResponse struct {
	ID                int                         `json:"id"`
	AccountID         int                         `json:"account_id"`
	Name              string                      `json:"name"`
	Address           string                      `json:"address"`
	City              string                      `json:"city"`
	Province          string                      `json:"province"`
	DefaultMeterItems []BuildingMeterItemResponse `json:"default_meter_items"`
}

// ToBuildingResponse 将楼栋领域对象转换为响应对象。
func ToBuildingResponse(building *domainbuilding.Building) BuildingResponse {
	if building == nil {
		return BuildingResponse{}
	}
	return BuildingResponse{
		ID:                building.ID,
		AccountID:         building.AccountID,
		Name:              building.Name,
		Address:           building.Address,
		City:              building.City,
		Province:          building.Province,
		DefaultMeterItems: toBuildingMeterItemResponses(building.DefaultMeterItems),
	}
}

// ToBuildingListResponse 将楼栋列表转换为响应列表。
func ToBuildingListResponse(buildings []*domainbuilding.Building) []BuildingResponse {
	ret := make([]BuildingResponse, 0, len(buildings))
	for _, building := range buildings {
		ret = append(ret, ToBuildingResponse(building))
	}
	return ret
}

// BuildingResponseRet 返回楼栋接口响应。
func BuildingResponseRet(building *domainbuilding.Building) *ginutils.Ret[BuildingResponse] {
	return ResponseWithData(ToBuildingResponse(building))
}

// BuildingListResponseRet 返回楼栋列表接口响应。
func BuildingListResponseRet(buildings []*domainbuilding.Building) *ginutils.Ret[[]BuildingResponse] {
	return ResponseWithData(ToBuildingListResponse(buildings))
}

func toBuildingMeterItemResponses(items []*domainbuilding.MeterItem) []BuildingMeterItemResponse {
	ret := make([]BuildingMeterItemResponse, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ret = append(ret, BuildingMeterItemResponse{
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
