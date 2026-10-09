package room

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrRoomNoRequired     = xerr.New(202001, "房间号不能为空")
	ErrInvalidFloor       = xerr.New(202002, "楼层不能小于 0")
	ErrInvalidPrice       = xerr.New(202003, "租金不能小于 0")
	ErrInvalidSize        = xerr.New(202004, "房间面积必须大于 0")
	ErrUnitRequired       = xerr.New(202005, "房间所属单元不能为空")
	ErrRoomTypeRequired   = xerr.New(202006, "房型不能为空")
	ErrInvalidMeterItemID = xerr.New(202007, "抄表项 ID 不合法")
	ErrMeterItemNotFound  = xerr.New(202008, "抄表项不存在")
	ErrNotFound           = xerr.New(202009, "房间不存在")
)
