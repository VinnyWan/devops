package middleware

import (
	"github.com/gin-gonic/gin"
)

func Casbin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以添加 Casbin RBAC 鉴权逻辑
		c.Next()
	}
}
