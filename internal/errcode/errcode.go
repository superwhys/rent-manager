package errcode

import (
	"errors"
	"fmt"
)

type Error struct {
	ErrCode int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) String() string {
	return e.Message
}

func (e Error) Code() int {
	return e.ErrCode
}

func (e Error) Is(target error) bool {
	switch t := target.(type) {
	case Error:
		return e.ErrCode != 0 && e.ErrCode == t.ErrCode
	case *Error:
		if t == nil {
			return false
		}
		return e.ErrCode != 0 && e.ErrCode == t.ErrCode
	default:
		return false
	}
}

func (e Error) WithErrCode(code int) Error {
	e.ErrCode = code
	return e
}

func (e Error) WithMessage(message string) Error {
	e.Message = message
	return e
}

func (e Error) AddMessage(message string) Error {
	e.Message = fmt.Sprintf("%s, %s", e.Message, message)
	return e
}

var (
	ErrSysInternal     = Error{ErrCode: 100500, Message: "系统内部错误"}
	ErrNotFound        = Error{ErrCode: 100404, Message: "资源不存在"}
	ErrBadRequest      = Error{ErrCode: 100400, Message: "请求参数错误"}
	ErrUnauthorized    = Error{ErrCode: 100401, Message: "未登录"}
	ErrForbidden       = Error{ErrCode: 100403, Message: "禁止访问"}
	ErrTooManyRequests = Error{ErrCode: 100429, Message: "请求过于频繁，请稍后再试"}

	ErrTokenExpired = Error{ErrCode: 400002, Message: "登录已过期"}
	ErrInvalidToken = Error{ErrCode: 400003, Message: "登录无效"}
)

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
