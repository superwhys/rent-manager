package meterreading

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrRoomRequired             = xerr.New(205001, "房间不能为空")
	ErrMeterItemRequired        = xerr.New(205002, "抄表项不能为空")
	ErrPeriodMonthRequired      = xerr.New(205003, "抄表月份不能为空")
	ErrInvalidPeriodMonth       = xerr.New(205004, "抄表月份必须使用 YYYY-MM 格式")
	ErrInvalidPreviousReading   = xerr.New(205005, "上次读数不能小于 0")
	ErrInvalidCurrentReading    = xerr.New(205006, "本次读数不能小于 0")
	ErrInvalidUsageAmount       = xerr.New(205007, "用量不能小于 0")
	ErrPreviousReadingRequired  = xerr.New(205008, "上次读数不能为空")
	ErrCurrentReadingDecreased  = xerr.New(205009, "本次读数不能小于上次读数")
	ErrReadingTimeRequired      = xerr.New(205010, "抄表时间不能为空")
	ErrInvalidSource            = xerr.New(205011, "抄表来源不合法")
	ErrInvalidCalcType          = xerr.New(205012, "抄表项计费类型不合法")
	ErrInvalidUnitPrice         = xerr.New(205013, "单价不能小于 0")
	ErrNotFound                 = xerr.New(205014, "抄表记录不存在")
	ErrAlreadyExists            = xerr.New(205015, "抄表记录已存在")
	ErrRoomNotFound             = xerr.New(205016, "房间不存在")
	ErrMeterItemNotFound        = xerr.New(205017, "抄表项不存在")
	ErrRoomMeterItemNotAssigned = xerr.New(205018, "房间未绑定该抄表项")
	ErrReadingsRequired         = xerr.New(205019, "抄表记录不能为空")
	ErrDuplicateMeterItem       = xerr.New(205020, "抄表项不能重复")
)
