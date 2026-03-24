package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"devops-platform/internal/pkg/logger"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("panic",
					zap.Any("panic", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", debug.Stack()),
				)
				c.JSON(500, gin.H{
					"message": "服务器内部错误",
					"error":   fmt.Sprintf("%v", r),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
