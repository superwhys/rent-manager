package meteritem

import "errors"

var (
	ErrAccountRequired   = errors.New("账号不能为空")
	ErrCodeRequired      = errors.New("抄表项编码不能为空")
	ErrUnitRequired      = errors.New("抄表项计量单位不能为空")
	ErrInvalidUnitPrice  = errors.New("抄表项单价不能小于 0")
	ErrInvalidCalcType   = errors.New("抄表项计费类型不合法")
	ErrNotFound          = errors.New("抄表项不存在")
	ErrCodeAlreadyExists = errors.New("抄表项编码已存在")
)
