package requests

type CreateRoomTypeCommand struct {
	Name string `json:"name" validate:"required,max=12"`
}

type UpdateRoomTypeCommand struct {
	RoomTypeID int    `uri:"room_type_id" validate:"required,min=1"`
	Name       string `json:"name" validate:"required,max=12"`
}

type ListRoomTypesQuery struct {
	Pagination

	Name string `form:"name"`
}

type RoomTypeIDQuery struct {
	RoomTypeID int `uri:"room_type_id" validate:"required,min=1"`
}
