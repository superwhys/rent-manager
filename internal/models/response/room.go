package response

import (
	domainroom "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/room"
	"github.com/miebyte/goutils/ginutils"
)

type RoomResponse struct {
	ID         int                 `json:"id"`
	RoomNo     string              `json:"room_no"`
	Floor      int                 `json:"floor"`
	Price      int                 `json:"price"`
	Size       int                 `json:"size"`
	UnitID     int                 `json:"unit_id"`
	RoomTypeID int                 `json:"room_type_id"`
	IsActive   bool                `json:"is_active"`
	MeterItems []MeterItemResponse `json:"meter_items"`
}

func ToRoomResponse(room *domainroom.Room) RoomResponse {
	return RoomResponse{
		ID:         room.ID,
		RoomNo:     room.RoomNo,
		Floor:      room.Floor,
		Price:      room.Price,
		Size:       room.Size,
		UnitID:     room.UnitID,
		RoomTypeID: room.RoomTypeID,
		IsActive:   room.IsActive,
		MeterItems: ToMeterItemListResponse(room.MeterItems),
	}
}

func ToRoomListResponse(rooms []*domainroom.Room) []RoomResponse {
	responses := make([]RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responses = append(responses, ToRoomResponse(room))
	}
	return responses
}

func RoomResponseRet(room *domainroom.Room) *ginutils.Ret[RoomResponse] {
	return ResponseWithData(ToRoomResponse(room))
}

func RoomListResponseRet(rooms []*domainroom.Room) *ginutils.Ret[[]RoomResponse] {
	return ResponseWithData(ToRoomListResponse(rooms))
}
