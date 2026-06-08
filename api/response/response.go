package response

import (
	"github.com/miebyte/goutils/ginutils"

	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// ResponseSuccess 返回无业务数据的成功响应。
func ResponseSuccess() *ginutils.Ret[any] {
	return ginutils.SuccessRet[any](nil)
}

// ResponseWithData 返回携带业务数据的成功响应。
func ResponseWithData[T any](data T) *ginutils.Ret[T] {
	return ginutils.SuccessRet(data)
}

// ErrorResponseWithCode 返回指定错误码对应的失败响应。
func ErrorResponseWithCode(ec xerr.Error) *ginutils.Ret[any] {
	return ginutils.ErrorRet(ec)
}
