package account

import "errors"

var (
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = errors.New("账号不存在")
	// ErrWechatOpenIDRequired 表示微信 openid 不能为空。
	ErrWechatOpenIDRequired = errors.New("微信 openid 不能为空")
)
