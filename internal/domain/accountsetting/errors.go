package accountsetting

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrAccountRequired = xerr.New(208001, "账号不能为空")
	ErrKeyRequired     = xerr.New(208002, "配置键不能为空")
	ErrValueRequired   = xerr.New(208003, "配置值不能为空")
	ErrNotFound        = xerr.New(208006, "账号配置不存在")
)
