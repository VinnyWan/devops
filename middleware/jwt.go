package middleware

import (
	"strings"

	"devops/common"
	"devops/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "请提供认证Token")
			return
		}

		// 验证Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			common.Unauthorized(c, "Token格式错误")
			return
		}

		// 解析Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			common.Unauthorized(c, "Token无效或已过期")
			return
		}

		// 将用户信息存储到上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint {
	if userId, exists := c.Get("userId"); exists {
		return userId.(uint)
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}
