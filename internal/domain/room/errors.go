package room

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	// ErrRoomNoRequired 表示房间号不能为空。
	ErrRoomNoRequired = xerr.New(202001, "房间号不能为空")
	// ErrInvalidFloor 表示楼层不能为负数。
	ErrInvalidFloor = xerr.New(202002, "楼层不能小于 0")
	// ErrInvalidPrice 表示租金不能为负数。
	ErrInvalidPrice = xerr.New(202003, "租金不能小于 0")
	// ErrInvalidSize 表示房间面积必须大于 0。
	ErrInvalidSize = xerr.New(202004, "房间面积必须大于 0")
	// ErrUnitRequired 表示房间必须归属单元。
	ErrUnitRequired = xerr.New(202005, "房间所属单元不能为空")
	// ErrRoomTypeRequired 表示房间必须归属房型。
	ErrRoomTypeRequired = xerr.New(202006, "房型不能为空")
	// ErrInvalidMeterItemID 表示抄表项 ID 非法。
	ErrInvalidMeterItemID = xerr.New(202007, "抄表项 ID 不合法")
	// ErrMeterItemNotFound 表示抄表项不存在或不属于当前账号。
	ErrMeterItemNotFound = xerr.New(202008, "抄表项不存在")
	// ErrNotFound 表示房间不存在。
	ErrNotFound = xerr.New(202009, "房间不存在")
)
