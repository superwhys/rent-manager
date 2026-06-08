package account

import "git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"

var (
	ErrAccountNotFound      = xerr.New(200001, "账号不存在")
	ErrWechatOpenIDRequired = xerr.New(200002, "微信 openid 不能为空")
)
