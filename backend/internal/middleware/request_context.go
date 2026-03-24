package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultTimezone   = "Asia/Shanghai"
	defaultLocale     = "zh-CN"
	defaultTimeFormat = "2006-01-02 15:04:05"
)

func pickHeaderOrDefault(c *gin.Context, key string, defaultValue string) string {
	val := strings.TrimSpace(c.GetHeader(key))
	if val == "" {
		return defaultValue
	}
	return val
}

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestAt := time.Now().UTC()
		timezone := pickHeaderOrDefault(c, "X-Timezone", defaultTimezone)
		locale := pickHeaderOrDefault(c, "X-Locale", defaultLocale)
		timeFormat := pickHeaderOrDefault(c, "X-Time-Format", defaultTimeFormat)
		c.Set("requestAt", requestAt)
		c.Set("timezone", timezone)
		c.Set("locale", locale)
		c.Set("time_format", timeFormat)
		if _, exists := c.Get("userID"); !exists {
			c.Set("userID", uint(0))
		}
		if _, exists := c.Get("username"); !exists {
			c.Set("username", "anonymous")
		}
		c.Next()
		c.Header("X-Request-Time", requestAt.Format(time.RFC3339))
		c.Header("X-Request-Timezone", timezone)
		c.Header("X-Request-Locale", locale)
		if username, ok := c.Get("username"); ok {
			if s, ok := username.(string); ok && s != "" {
				c.Header("X-Request-User", s)
			}
		}
	}
}
