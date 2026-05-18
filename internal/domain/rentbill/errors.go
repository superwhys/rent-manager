package rentbill

import "errors"

var (
	ErrRoomRequired        = errors.New("房间不能为空")
	ErrPeriodMonthRequired = errors.New("账单月份不能为空")
	ErrInvalidPeriodMonth  = errors.New("账单月份必须使用 YYYY-MM 格式")
	ErrInvalidBaseRent     = errors.New("基础租金不能小于 0")
	ErrInvalidMeterFee     = errors.New("抄表费用不能小于 0")
	ErrInvalidTotalAmount  = errors.New("账单总金额不能小于 0")
	ErrInvalidStatus       = errors.New("租金账单状态不合法")
	ErrPaidAtRequired      = errors.New("支付时间不能为空")
	ErrAlreadyPaid         = errors.New("租金账单已支付")
	ErrCancelled           = errors.New("租金账单已取消")
	ErrNotFound            = errors.New("租金账单不存在")
	ErrAlreadyExists       = errors.New("租金账单已存在")
	ErrRoomNotFound        = errors.New("房间不存在")
)
