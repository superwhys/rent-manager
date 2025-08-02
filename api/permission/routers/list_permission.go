package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
)

type ListPermissionHandler struct {
}

func (h ListPermissionHandler) Handle(c *gin.Context, app App) (*ginutils.Ret, error) {
	return &ginutils.Ret{
		Code:    0,
		Message: "success",
		Data:    gin.H{"message": "Hello, World!"},
	}, nil
}
