package room

import (
	appcore "github.com/superwhys/rent-manager/internal/app"
	appmeteritem "github.com/superwhys/rent-manager/internal/app/meteritem"
	domainroom "github.com/superwhys/rent-manager/internal/domain/room"
)

// CreateRoomCommand 表示创建房间命令。
type CreateRoomCommand struct {
	RoomNo       string `json:"room_no" validate:"required"`
	Floor        int    `json:"floor"`
	Price        int    `json:"price" validate:"required,min=1"`
	Size         int    `json:"size"`
	UnitID       int    `json:"unit_id" validate:"required,min=1"`
	RoomTypeID   int    `json:"room_type_id" validate:"required,min=1"`
	MeterItemIDs []int  `json:"meter_item_ids" validate:"omitempty,dive,min=1"`
}

// CreateRoomInUnitCommand 是在单元下创建房间的请求参数。
type CreateRoomInUnitCommand struct {
	UnitID       int    `uri:"unit_id" validate:"required,min=1"`
	RoomNo       string `json:"room_no" validate:"required"`
	Floor        int    `json:"floor"`
	Price        int    `json:"price" validate:"required,min=1"`
	Size         int    `json:"size" validate:"omitempty,min=1"`
	RoomTypeID   int    `json:"room_type_id" validate:"required,min=1"`
	MeterItemIDs []int  `json:"meter_item_ids" validate:"omitempty,dive,min=1"`
}

// UpdateRoomCommand 表示更新房间命令。
type UpdateRoomCommand struct {
	RoomID       int    `uri:"room_id" validate:"required,min=1"`
	RoomNo       string `json:"room_no"`
	Floor        *int   `json:"floor" validate:"omitempty,min=0"`
	Price        *int   `json:"price" validate:"omitempty,min=0"`
	Size         *int   `json:"size" validate:"omitempty,min=1"`
	UnitID       *int   `json:"unit_id" validate:"omitempty,min=1"`
	RoomTypeID   *int   `json:"room_type_id" validate:"omitempty,min=1"`
	IsActive     *bool  `json:"is_active"`
	MeterItemIDs []int  `json:"meter_item_ids" validate:"omitempty,dive,min=1"`
}

// ListRoomsQuery 表示房间列表查询参数。
type ListRoomsQuery struct {
	appcore.Pagination

	UnitID     int   `form:"unit_id" validate:"required,min=1"`
	RoomTypeID *int  `form:"room_type_id"`
	IsActive   *bool `form:"is_active"`
}

// RoomIDQuery 表示房间 ID 路径参数。
type RoomIDQuery struct {
	RoomID int `uri:"room_id" validate:"required,min=1"`
}

// UnitRoomsQuery 表示单元下房间列表查询参数。
type UnitRoomsQuery struct {
	appcore.Pagination

	UnitID     int   `uri:"unit_id" validate:"required,min=1"`
	RoomTypeID *int  `form:"room_type_id"`
	IsActive   *bool `form:"is_active"`
}

// RoomResponse 是房间响应。
type RoomResponse struct {
	ID         int                              `json:"id"`
	RoomNo     string                           `json:"room_no"`
	Floor      int                              `json:"floor"`
	Price      int                              `json:"price"`
	Size       int                              `json:"size"`
	UnitID     int                              `json:"unit_id"`
	RoomTypeID int                              `json:"room_type_id"`
	IsActive   bool                             `json:"is_active"`
	MeterItems []appmeteritem.MeterItemResponse `json:"meter_items"`
}

// ToRoomResponse 将房间领域对象转换为响应对象。
func ToRoomResponse(room *domainroom.Room) RoomResponse {
	if room == nil {
		return RoomResponse{}
	}
	return RoomResponse{
		ID:         room.ID,
		RoomNo:     room.RoomNo,
		Floor:      room.Floor,
		Price:      room.Price,
		Size:       room.Size,
		UnitID:     room.UnitID,
		RoomTypeID: room.RoomTypeID,
		IsActive:   room.IsActive,
		MeterItems: appmeteritem.ToMeterItemListResponse(room.MeterItems),
	}
}

// ToRoomListResponse 将房间列表转换为响应列表。
func ToRoomListResponse(rooms []*domainroom.Room) []RoomResponse {
	responses := make([]RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responses = append(responses, ToRoomResponse(room))
	}
	return responses
}
