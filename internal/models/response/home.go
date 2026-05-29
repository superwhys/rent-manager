package response

import "github.com/miebyte/goutils/ginutils"

// HomeBuildingRoomSummaryResponse 是首页楼栋与房间汇总响应。
type HomeBuildingRoomSummaryResponse struct {
	BuildingCount   int `json:"building_count"`
	RoomCount       int `json:"room_count"`
	RentedRoomCount int `json:"rented_room_count"`
	VacantRoomCount int `json:"vacant_room_count"`
}

// HomeStatResponse 是首页核心统计项响应。
type HomeStatResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

// HomeTodoResponse 是首页待办项响应。
type HomeTodoResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Tag         string `json:"tag"`
}

// HomeResponse 是首页聚合数据响应。
type HomeResponse struct {
	BuildingRoomSummary HomeBuildingRoomSummaryResponse `json:"building_room_summary"`
	IncomeAmount        string                          `json:"incomeAmount"`
	IncomeDecimal       string                          `json:"incomeDecimal"`
	MonthlyGrowth       string                          `json:"monthlyGrowth"`
	CurrentMonth        string                          `json:"currentMonth"`
	Stats               []HomeStatResponse              `json:"stats"`
	Todos               []HomeTodoResponse              `json:"todos"`
}

// HomeResponseRet 返回首页聚合数据接口响应。
func HomeResponseRet(data HomeResponse) *ginutils.Ret[HomeResponse] {
	return ResponseWithData(data)
}
