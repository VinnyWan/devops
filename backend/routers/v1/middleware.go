package v1

import (
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

// authMiddlewares 鉴权中间件链
func authMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.SessionAuth(), // Session 认证
	}
}
