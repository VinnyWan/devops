package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

// SetDB 设置数据库连接
func SetDB(database *gorm.DB) {
	db = database
}

// GetCurrentUserID 从上下文中安全获取当前用户ID
func GetCurrentUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// GetCurrentUsername 从上下文中安全获取当前用户名
func GetCurrentUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}
