package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"devops-platform/internal/pkg/logger"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Log.Warn("failed to read request body", zap.Error(err))
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// response body
		bw := &bodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = bw

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		maskedBody := maskSensitiveFields(bodyBytes)

		level := zap.InfoLevel
		if status >= 500 {
			level = zap.ErrorLevel
		} else if status >= 400 {
			level = zap.WarnLevel
		}

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Any("query", c.Request.URL.Query()),
			zap.String("body", maskedBody),
			zap.Int("status", status),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("user", c.GetString("username")),
		}

		// 记录错误
		if len(c.Errors) > 0 {
			fields = append(fields, zap.Any("error", c.Errors.String()))
		}

		// 构造包含参数的 msg
		msg := "http_request"
		if len(maskedBody) > 0 {
			// 如果 body 太长，截断一下
			if len(maskedBody) > 500 {
				maskedBody = maskedBody[:500] + "..."
			}
			msg = "http_request: " + maskedBody
		} else if len(c.Request.URL.RawQuery) > 0 {
			msg = "http_request: " + c.Request.URL.RawQuery
		}

		logger.Log.Check(level, msg).Write(fields...)
	}
}
