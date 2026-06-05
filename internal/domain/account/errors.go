package account

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	// ErrAccountNotFound 表示账号不存在。
	ErrAccountNotFound = xerr.New(200001, "账号不存在")
	// ErrWechatOpenIDRequired 表示微信 openid 不能为空。
	ErrWechatOpenIDRequired = xerr.New(200002, "微信 openid 不能为空")
)
