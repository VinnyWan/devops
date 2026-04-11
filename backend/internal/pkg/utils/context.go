package utils

import (
	"github.com/gin-gonic/gin"
)

// GetCurrentTenantID 从上下文获取当前租户ID
func GetCurrentTenantID(c *gin.Context) uint {
	tenantID, exists := c.Get("tenantID")
	if !exists {
		return 0
	}
	if id, ok := tenantID.(uint); ok {
		return id
	}
	return 0
}

// GetCurrentTenantCode 从上下文获取当前租户代码
func GetCurrentTenantCode(c *gin.Context) string {
	tenantCode, exists := c.Get("tenantCode")
	if !exists {
		return ""
	}
	if code, ok := tenantCode.(string); ok {
		return code
	}
	return ""
}

// GetCurrentUserID 从上下文获取当前用户ID
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
