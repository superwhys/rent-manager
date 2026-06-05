package unit

import domainunit "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/unit"

// CreateUnitCommand 是创建单元的请求体。
type CreateUnitCommand struct {
	Name                string `json:"name" validate:"required,max=255"`
	Address             string `json:"address" validate:"required,max=255"`
	City                string `json:"city" validate:"required,max=255"`
	Province            string `json:"province" validate:"required,max=255"`
	DefaultMeterItemIDs []int  `json:"default_meter_item_ids" validate:"omitempty,dive,min=1"`
}

// UpdateUnitCommand 是更新单元的请求体。
type UpdateUnitCommand struct {
	UnitID              int    `uri:"unit_id" validate:"required,min=1"`
	Name                string `json:"name" validate:"required,max=255"`
	Address             string `json:"address" validate:"required,max=255"`
	City                string `json:"city" validate:"required,max=255"`
	Province            string `json:"province" validate:"required,max=255"`
	DefaultMeterItemIDs []int  `json:"default_meter_item_ids" validate:"omitempty,dive,min=1"`
}

// ListUnitsQuery 是单元列表查询参数。
type ListUnitsQuery struct {
	Scope string `form:"scope" validate:"omitempty,oneof=owned granted all"`
}

// UnitIDQuery 是单元 ID 路径参数。
type UnitIDQuery struct {
	UnitID int `uri:"unit_id" validate:"required,min=1"`
}

// UnitMeterItemResponse 是单元默认抄表项响应。
type UnitMeterItemResponse struct {
	ID        int     `json:"id"`
	Code      string  `json:"code"`
	Unit      string  `json:"unit"`
	UnitPrice int     `json:"unit_price"`
	IsSystem  bool    `json:"is_system"`
	CalcType  string  `json:"calc_type"`
	Remark    *string `json:"remark,omitempty"`
}

// UnitResponse 是单元响应。
type UnitResponse struct {
	ID                int                     `json:"id"`
	AccountID         int                     `json:"account_id"`
	Name              string                  `json:"name"`
	Address           string                  `json:"address"`
	City              string                  `json:"city"`
	Province          string                  `json:"province"`
	AccessType        string                  `json:"access_type"`
	IsOwner           bool                    `json:"is_owner"`
	DefaultMeterItems []UnitMeterItemResponse `json:"default_meter_items"`
	RoomCount         int                     `json:"room_count"`
	RentedCount       int                     `json:"rented_count"`
	MonthlyRentAmount string                  `json:"monthly_rent_amount"`
	OccupancyStatus   string                  `json:"occupancy_status"`
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
		AccessType:        unit.AccessType,
		IsOwner:           unit.AccessType == domainunit.AccessTypeOwned,
		DefaultMeterItems: toUnitMeterItemResponses(unit.DefaultMeterItems),
	}
}

// ToUnitSummaryResponse 将单元统计对象转换为响应对象。
func ToUnitSummaryResponse(summary *domainunit.Summary) UnitResponse {
	if summary == nil {
		return UnitResponse{}
	}
	response := ToUnitResponse(summary.Unit)
	response.RoomCount = summary.RoomCount
	response.RentedCount = summary.RentedCount
	response.MonthlyRentAmount = summary.MonthlyRentAmount
	response.OccupancyStatus = summary.OccupancyStatus
	return response
}

// ToUnitListResponse 将单元统计列表转换为响应列表。
func ToUnitListResponse(summaries []*domainunit.Summary) []UnitResponse {
	ret := make([]UnitResponse, 0, len(summaries))
	for _, summary := range summaries {
		ret = append(ret, ToUnitSummaryResponse(summary))
	}
	return ret
}

// toUnitMeterItemResponses 将单元默认抄表项转换为响应列表。
func toUnitMeterItemResponses(items []*domainunit.MeterItem) []UnitMeterItemResponse {
	ret := make([]UnitMeterItemResponse, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ret = append(ret, UnitMeterItemResponse{
			ID:        item.ID,
			Code:      item.Code,
			Unit:      item.Unit,
			UnitPrice: item.UnitPrice,
			IsSystem:  item.IsSystem,
			CalcType:  item.CalcType,
			Remark:    item.Remark,
		})
	}
	return ret
}
