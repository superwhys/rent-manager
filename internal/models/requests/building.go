package requests

// CreateBuildingCommand 是创建楼栋的请求体。
type CreateBuildingCommand struct {
	OrganizationID      int    `json:"organization_id" validate:"required,min=1"`
	Name                string `json:"name" validate:"required,max=255"`
	Address             string `json:"address" validate:"required,max=255"`
	City                string `json:"city" validate:"required,max=255"`
	Province            string `json:"province" validate:"required,max=255"`
	DefaultMeterItemIDs []int  `json:"default_meter_item_ids" validate:"omitempty,dive,min=1"`
}

// UpdateBuildingCommand 是更新楼栋的请求体。
type UpdateBuildingCommand struct {
	OrganizationID      int    `json:"organization_id" validate:"required,min=1"`
	BuildingID          int    `uri:"building_id" validate:"required,min=1"`
	Name                string `json:"name" validate:"required,max=255"`
	Address             string `json:"address" validate:"required,max=255"`
	City                string `json:"city" validate:"required,max=255"`
	Province            string `json:"province" validate:"required,max=255"`
	DefaultMeterItemIDs []int  `json:"default_meter_item_ids" validate:"omitempty,dive,min=1"`
}

// ListBuildingsQuery 是楼栋列表查询参数。
type ListBuildingsQuery struct {
	Pagination

	OrganizationID int    `form:"organization_id" validate:"required,min=1"`
	Name           string `form:"name"`
	City           string `form:"city"`
	Province       string `form:"province"`
}

// BuildingIDQuery 是楼栋 ID 路径参数。
type BuildingIDQuery struct {
	OrganizationID int `form:"organization_id" validate:"required,min=1"`
	BuildingID     int `uri:"building_id" validate:"required,min=1"`
}

// CreateRoomInBuildingCommand 是在楼栋下创建房间的请求参数。
type CreateRoomInBuildingCommand struct {
	OrganizationID int    `json:"organization_id" validate:"required,min=1"`
	BuildingID     int    `uri:"building_id" validate:"required,min=1"`
	RoomNo         string `json:"room_no" validate:"required"`
	Floor          int    `json:"floor"`
	Price          int    `json:"price" validate:"required,min=1"`
	Size           int    `json:"size" validate:"required,min=1"`
	RoomTypeID     int    `json:"room_type_id" validate:"required,min=1"`
	MeterItemIDs   []int  `json:"meter_item_ids" validate:"omitempty,dive,min=1"`
}
