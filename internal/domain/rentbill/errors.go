package rentbill

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrRoomRequired        = xerr.New(206001, "房间不能为空")
	ErrPeriodMonthRequired = xerr.New(206002, "账单月份不能为空")
	ErrInvalidPeriodMonth  = xerr.New(206003, "账单月份必须使用 YYYY-MM 格式")
	ErrInvalidBaseRent     = xerr.New(206004, "基础租金不能小于 0")
	ErrInvalidMeterFee     = xerr.New(206005, "抄表费用不能小于 0")
	ErrInvalidTotalAmount  = xerr.New(206006, "账单总金额不能小于 0")
	ErrInvalidStatus       = xerr.New(206007, "租金账单状态不合法")
	ErrPaidAtRequired      = xerr.New(206008, "支付时间不能为空")
	ErrAlreadyPaid         = xerr.New(206009, "租金账单已支付")
	ErrCancelled           = xerr.New(206010, "租金账单已取消")
	ErrNotFound            = xerr.New(206011, "租金账单不存在")
	ErrAlreadyExists       = xerr.New(206012, "租金账单已存在")
	ErrRoomNotFound        = xerr.New(206013, "房间不存在")
	ErrBillsRequired       = xerr.New(206014, "租金账单不能为空")
	ErrMeterReadingMissing = xerr.New(206015, "抄表未完成，不能生成账单")
)
