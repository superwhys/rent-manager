package meterreading

import "errors"

var (
	ErrRoomRequired             = errors.New("房间不能为空")
	ErrMeterItemRequired        = errors.New("抄表项不能为空")
	ErrPeriodMonthRequired      = errors.New("抄表月份不能为空")
	ErrInvalidPeriodMonth       = errors.New("抄表月份必须使用 YYYY-MM 格式")
	ErrInvalidPreviousReading   = errors.New("上次读数不能小于 0")
	ErrInvalidCurrentReading    = errors.New("本次读数不能小于 0")
	ErrInvalidUsageAmount       = errors.New("用量不能小于 0")
	ErrPreviousReadingRequired  = errors.New("上次读数不能为空")
	ErrCurrentReadingDecreased  = errors.New("本次读数不能小于上次读数")
	ErrReadingTimeRequired      = errors.New("抄表时间不能为空")
	ErrInvalidSource            = errors.New("抄表来源不合法")
	ErrInvalidCalcType          = errors.New("抄表项计费类型不合法")
	ErrInvalidUnitPrice         = errors.New("单价不能小于 0")
	ErrNotFound                 = errors.New("抄表记录不存在")
	ErrAlreadyExists            = errors.New("抄表记录已存在")
	ErrRoomNotFound             = errors.New("房间不存在")
	ErrMeterItemNotFound        = errors.New("抄表项不存在")
	ErrRoomMeterItemNotAssigned = errors.New("房间未绑定该抄表项")
)
