package middleware

import (
	"net/http"

	"devops-platform/internal/modules/user/service"

	"github.com/gin-gonic/gin"
)

// RequirePermission 权限校验中间件
// resource: 资源名称
// action: 操作名称
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前用户ID
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：无法获取用户信息",
			})
			return
		}
		userID := userIDVal.(uint)

		// 2. 校验权限
		if db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "系统错误：数据库连接未初始化",
			})
			return
		}

		userSvc := service.NewUserService(db)
		// 使用 CheckPermission
		allowed, err := userSvc.CheckPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限校验失败: " + err.Error(),
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			return
		}

		c.Next()
	}
}
