package room

import "errors"

var (
	// ErrRoomNoRequired 表示房间号不能为空。
	ErrRoomNoRequired = errors.New("房间号不能为空")
	// ErrInvalidFloor 表示楼层不能为负数。
	ErrInvalidFloor = errors.New("楼层不能小于 0")
	// ErrInvalidPrice 表示租金不能为负数。
	ErrInvalidPrice = errors.New("租金不能小于 0")
	// ErrInvalidSize 表示房间面积必须大于 0。
	ErrInvalidSize = errors.New("房间面积必须大于 0")
	// ErrUnitRequired 表示房间必须归属单元。
	ErrUnitRequired = errors.New("房间所属单元不能为空")
	// ErrRoomTypeRequired 表示房间必须归属房型。
	ErrRoomTypeRequired = errors.New("房型不能为空")
	// ErrInvalidMeterItemID 表示抄表项 ID 非法。
	ErrInvalidMeterItemID = errors.New("抄表项 ID 不合法")
	// ErrMeterItemNotFound 表示抄表项不存在或不属于当前账号。
	ErrMeterItemNotFound = errors.New("抄表项不存在")
	// ErrNotFound 表示房间不存在。
	ErrNotFound = errors.New("房间不存在")
)
