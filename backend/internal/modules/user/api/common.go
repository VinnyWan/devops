package api

import (
	"errors"
	"net/http"
	"strings"
	"devops-platform/internal/modules/user/model"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

// SetDB 设置数据库连接
func SetDB(database *gorm.DB) {
	db = database
	authService = nil
	authOnce = sync.Once{}
	userService = nil
	userOnce = sync.Once{}
	deptService = nil
	deptOnce = sync.Once{}
	deptUserService = nil
	deptUserOnce = sync.Once{}
	roleService = nil
	roleOnce = sync.Once{}
	auditService = nil
	auditOnce = sync.Once{}
	tenantService = nil
	tenantOnce = sync.Once{}
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

// GetCurrentTenantID 从上下文中安全获取当前租户ID
func GetCurrentTenantID(c *gin.Context) uint {
	tenantID, exists := c.Get("tenantID")
	if !exists {
		// 兼容旧测试/调用：未注入 tenantID 时，尝试通过 userID 回查
		userID := GetCurrentUserID(c)
		if userID == 0 || db == nil {
			return 0
		}
		var user model.User
		if err := db.Select("tenant_id").Where("id = ?", userID).First(&user).Error; err == nil && user.TenantID != nil {
			return *user.TenantID
		}
		return 0
	}
	if id, ok := tenantID.(uint); ok {
		return id
	}
	return 0
}

// GetCurrentTenantCode 从上下文中安全获取当前租户编码
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

func writeModuleError(c *gin.Context, err error, fallbackStatus int) {
	status := fallbackStatus
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
	case strings.Contains(err.Error(), "permission denied"):
		status = http.StatusForbidden
	}

	c.JSON(status, gin.H{
		"code":    status,
		"message": err.Error(),
	})
}
