package xerr

import "errors"

var (
	// ErrSysInternal 表示系统内部错误。
	ErrSysInternal = New(500001, "系统内部错误")
	// ErrBadRequest 表示请求参数不符合业务要求。
	ErrBadRequest = New(400001, "请求参数错误")
	// ErrUnauthorized 表示当前请求缺少登录态。
	ErrUnauthorized = New(400002, "未登录")
	// ErrTokenExpired 表示访问令牌已过期。
	ErrTokenExpired = New(400003, "令牌已过期")
	// ErrLoginExpired 表示登录态已过期。
	ErrLoginExpired = New(400003, "登录已过期")
	// ErrInvalidToken 表示令牌无效。
	ErrInvalidToken = New(400004, "无效令牌")
	// ErrTooManyRequests 表示请求过于频繁。
	ErrTooManyRequests = New(429001, "请求过于频繁")
)

// IsAuthExpired 判断错误是否表示认证过期。
func IsAuthExpired(err error) bool {
	return errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrLoginExpired)
}

// AsErrcode 从错误链中提取业务错误。
func AsErrcode(err error) (Error, bool) {
	if err == nil {
		return Error{}, false
	}

	ec := new(Error)
	if errors.As(err, ec) {
		return *ec, true
	}

	return Error{}, false
}
