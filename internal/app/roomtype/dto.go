package roomtype

import (
	appcore "github.com/superwhys/rent-manager/internal/app"
	domainroomtype "github.com/superwhys/rent-manager/internal/domain/roomtype"
)

// CreateRoomTypeCommand 表示创建房型命令。
type CreateRoomTypeCommand struct {
	Name string `json:"name" validate:"required,max=12"`
}

// UpdateRoomTypeCommand 表示更新房型命令。
type UpdateRoomTypeCommand struct {
	RoomTypeID int    `uri:"room_type_id" validate:"required,min=1"`
	Name       string `                   validate:"required,max=12" json:"name"`
}

// ListRoomTypesQuery 表示房型列表查询参数。
type ListRoomTypesQuery struct {
	appcore.Pagination

	Name string `form:"name"`
}

// RoomTypeIDQuery 表示房型 ID 路径参数。
type RoomTypeIDQuery struct {
	RoomTypeID int `uri:"room_type_id" validate:"required,min=1"`
}

// RoomTypeResponse 是房型响应。
type RoomTypeResponse struct {
	ID        int    `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
}

// ToRoomTypeResponse 将房型领域对象转换为响应对象。
func ToRoomTypeResponse(roomType *domainroomtype.RoomType) RoomTypeResponse {
	if roomType == nil {
		return RoomTypeResponse{}
	}
	return RoomTypeResponse{
		ID:        roomType.ID,
		AccountID: roomType.AccountID,
		Name:      roomType.Name,
	}
}

// ToRoomTypeListResponse 将房型列表转换为响应列表。
func ToRoomTypeListResponse(roomTypes []*domainroomtype.RoomType) []RoomTypeResponse {
	responses := make([]RoomTypeResponse, 0, len(roomTypes))
	for _, roomType := range roomTypes {
		responses = append(responses, ToRoomTypeResponse(roomType))
	}
	return responses
}
