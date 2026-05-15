package account

import "errors"

var (
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = errors.New("account not found")
)
