package middleware

import (
	"fmt"
	"net/http"

	"devops-platform/internal/modules/user/service"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

// casbinEnforcer 包级 Casbin enforcer 引用，由 SetCasbinEnforcer 注入
var casbinEnforcer *casbin.SyncedEnforcer

// SetCasbinEnforcer 设置 Casbin enforcer 实例，供 bootstrap 初始化时调用
func SetCasbinEnforcer(enforcer *casbin.SyncedEnforcer) {
	casbinEnforcer = enforcer
}

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
		userID, ok := userIDVal.(uint)
		if !ok || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：用户信息无效",
			})
			return
		}
		tenantIDVal, exists := c.Get("tenantID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：无法获取租户信息",
			})
			return
		}
		tenantID, ok := tenantIDVal.(uint)
		if !ok || tenantID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证：租户信息无效或缺失",
			})
			return
		}

		// 2. Casbin 优先路径：如果 enforcer 已注入，优先使用 Casbin Enforce
		if casbinEnforcer != nil {
			sub := fmt.Sprintf("%d:%d", tenantID, userID)
			allowed, err := casbinEnforcer.Enforce(sub, resource, action)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Casbin 权限校验失败: " + err.Error(),
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
			return
		}

		// 3. 降级到原有自建权限校验逻辑
		if db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "系统错误：数据库连接未初始化",
			})
			return
		}

		userSvc := service.NewUserService(db)
		// 使用 CheckPermission
		allowed, err := userSvc.CheckPermission(c.Request.Context(), tenantID, userID, resource, action)
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
