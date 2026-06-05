package xerr

import (
	"fmt"
)

// Error 表示跨层传递的业务错误。
type Error struct {
	ErrCode int
	Message string
}

// New 创建业务错误。
func New(code int, message string) Error {
	return Error{ErrCode: code, Message: message}
}

// Error 返回错误消息。
func (e Error) Error() string {
	return e.Message
}

// String 返回错误消息。
func (e Error) String() string {
	return e.Message
}

// Code 返回业务错误码。
func (e Error) Code() int {
	return e.ErrCode
}

// Is 根据错误码判断两个业务错误是否相同。
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

// WithErrCode 返回替换业务错误码后的错误。
func (e Error) WithErrCode(code int) Error {
	e.ErrCode = code
	return e
}

// WithMessage 返回替换错误消息后的错误。
func (e Error) WithMessage(message string) Error {
	e.Message = message
	return e
}

// AddMessage 返回追加错误消息后的错误。
func (e Error) AddMessage(message string) Error {
	e.Message = fmt.Sprintf("%s, %s", e.Message, message)
	return e
}
