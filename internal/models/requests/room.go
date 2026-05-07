package requests

type CreateRoomCommand struct {
	RoomNo     string `json:"room_no" validate:"required"`
	Floor      int    `json:"floor"`
	Price      int    `json:"price" validate:"required,min=1"`
	Size       int    `json:"size" `
	BuildingID int    `json:"building_id" validate:"required,min=1"`
	RoomTypeID int    `json:"room_type_id" validate:"required,min=1"`
}

type UpdateRoomCommand struct {
	RoomID     int    `uri:"room_id" validate:"required,min=1"`
	RoomNo     string `json:"room_no"`
	Floor      int    `json:"floor"`
	Price      int    `json:"price"`
	Size       int    `json:"size"`
	BuildingID int    `json:"building_id"`
	RoomTypeID int    `json:"room_type_id"`
	IsActive   bool   `json:"is_active"`
}

type ListRoomsQuery struct {
	Pagination

	BuildingID int   `form:"building_id" validate:"required,min=1"`
	RoomTypeID *int  `form:"room_type_id"`
	IsActive   *bool `form:"is_active"`
}

type RoomIdQuery struct {
	RoomID int `uri:"room_id" validate:"required,min=1"`
}
