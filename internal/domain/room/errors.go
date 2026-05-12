package room

import "errors"

var (
	// ErrRoomNoRequired 表示房间号不能为空。
	ErrRoomNoRequired = errors.New("room no is required")
	// ErrInvalidFloor 表示楼层不能为负数。
	ErrInvalidFloor = errors.New("floor must be greater than or equal to 0")
	// ErrInvalidPrice 表示租金不能为负数。
	ErrInvalidPrice = errors.New("price must be greater than or equal to 0")
	// ErrInvalidSize 表示房间面积必须大于 0。
	ErrInvalidSize = errors.New("size must be greater than 0")
	// ErrBuildingRequired 表示房间必须归属楼栋。
	ErrBuildingRequired = errors.New("building is required")
	// ErrRoomTypeRequired 表示房间必须归属房型。
	ErrRoomTypeRequired = errors.New("room type is required")
	// ErrInvalidMeterItemID 表示抄表项 ID 非法。
	ErrInvalidMeterItemID = errors.New("invalid meter item id")
	// ErrMeterItemNotFound 表示抄表项不存在或不属于当前账号。
	ErrMeterItemNotFound = errors.New("meter item not found")
	// ErrNotFound 表示房间不存在。
	ErrNotFound = errors.New("room not found")
)
