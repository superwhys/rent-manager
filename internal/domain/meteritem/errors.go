package meteritem

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	ErrAccountRequired   = xerr.New(204001, "账号不能为空")
	ErrCodeRequired      = xerr.New(204002, "抄表项编码不能为空")
	ErrUnitRequired      = xerr.New(204003, "抄表项计量单位不能为空")
	ErrInvalidUnitPrice  = xerr.New(204004, "抄表项单价不能小于 0")
	ErrInvalidCalcType   = xerr.New(204005, "抄表项计费类型不合法")
	ErrNotFound          = xerr.New(204006, "抄表项不存在")
	ErrCodeAlreadyExists = xerr.New(204007, "抄表项编码已存在")
)
