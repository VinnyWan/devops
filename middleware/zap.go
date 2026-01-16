package middleware

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// GinZapLogger 中间件：记录所有 HTTP 请求日志（JSON）
func GinZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		// 使用自定义的responseWriter来捕获响应内容
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		c.Next() // 处理请求

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 尝试从响应体中提取业务code
		businessCode := statusCode
		if blw.body.Len() > 0 {
			var response map[string]interface{}
			if err := json.Unmarshal(blw.body.Bytes(), &response); err == nil {
				if code, ok := response["code"]; ok {
					switch v := code.(type) {
					case float64:
						businessCode = int(v)
					case int:
						businessCode = v
					}
				}
			}
		}

		logger.Info("HTTP Request",
			zap.Int("status", businessCode), // 记录业务code而不是HTTP状态码
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
