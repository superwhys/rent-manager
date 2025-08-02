package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/rent-manager-server/model"
)

type RegisterHandler struct {
	model.RegisterRequest
}

func (h RegisterHandler) Handle(c *gin.Context, app App) (*ginutils.Ret, error) {
	return &ginutils.Ret{
		Code:    0,
		Message: "success",
		Data:    gin.H{"message": "Hello, World!"},
	}, nil
}
