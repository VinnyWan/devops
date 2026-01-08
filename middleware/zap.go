package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinZapLogger 中间件：记录所有 HTTP 请求日志（JSON）
func GinZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next() // 处理请求

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logger.Info("HTTP Request",
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("errors", errs),
		)
	}
}

// GinRecoveryWithZap 中间件：捕获 panic，并用 zap 输出 JSON 日志
func GinRecoveryWithZap(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic Recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
