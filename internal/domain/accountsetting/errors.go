package accountsetting

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	ErrAccountRequired        = xerr.New(208001, "账号不能为空")
	ErrKeyRequired            = xerr.New(208002, "配置键不能为空")
	ErrValueRequired          = xerr.New(208003, "配置值不能为空")
	ErrUnsupportedKey         = xerr.New(208004, "配置项不支持")
	ErrInvalidMeterReadingDay = xerr.New(208005, "抄表日必须是 1-31 的整数")
	ErrNotFound               = xerr.New(208006, "账号配置不存在")
)
