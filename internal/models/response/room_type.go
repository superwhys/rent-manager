package response

import (
	"github.com/miebyte/goutils/ginutils"
	domainroomtype "github.com/superwhys/rent-manager-server/internal/domain/roomtype"
)

type RoomTypeResponse struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Name      string `json:"name"`
}

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

func ToRoomTypeListResponse(roomTypes []*domainroomtype.RoomType) []RoomTypeResponse {
	responses := make([]RoomTypeResponse, 0, len(roomTypes))
	for _, roomType := range roomTypes {
		responses = append(responses, ToRoomTypeResponse(roomType))
	}
	return responses
}

func RoomTypeResponseRet(roomType *domainroomtype.RoomType) *ginutils.Ret[RoomTypeResponse] {
	return ResponseWithData(ToRoomTypeResponse(roomType))
}

func RoomTypeListResponseRet(roomTypes []*domainroomtype.RoomType) *ginutils.Ret[[]RoomTypeResponse] {
	return ResponseWithData(ToRoomTypeListResponse(roomTypes))
}
