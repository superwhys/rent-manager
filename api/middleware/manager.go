package middleware

import "github.com/gin-gonic/gin"

type Manager interface {
	Auth() gin.HandlerFunc
}

type manager struct {
}

func NewManager() Manager {
	return &manager{}
}

func (m *manager) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO:
		c.Next()
	}
}
