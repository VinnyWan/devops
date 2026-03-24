package middleware

import "github.com/gin-gonic/gin"

// Charset 统一设置响应字符集为UTF-8
func Charset() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置响应头字符集为UTF-8
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	}
}
