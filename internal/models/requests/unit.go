package requests

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
	Pagination

	Name     string `form:"name"`
	City     string `form:"city"`
	Province string `form:"province"`
	Scope    string `form:"scope" validate:"omitempty,oneof=owned granted all"`
}

// UnitIDQuery 是单元 ID 路径参数。
type UnitIDQuery struct {
	UnitID int `uri:"unit_id" validate:"required,min=1"`
}

// CreateRoomInUnitCommand 是在单元下创建房间的请求参数。
type CreateRoomInUnitCommand struct {
	UnitID       int    `uri:"unit_id" validate:"required,min=1"`
	RoomNo       string `json:"room_no" validate:"required"`
	Floor        int    `json:"floor"`
	Price        int    `json:"price" validate:"required,min=1"`
	Size         int    `json:"size" validate:"required,min=1"`
	RoomTypeID   int    `json:"room_type_id" validate:"required,min=1"`
	MeterItemIDs []int  `json:"meter_item_ids" validate:"omitempty,dive,min=1"`
}
