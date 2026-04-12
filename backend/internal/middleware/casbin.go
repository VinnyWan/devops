package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// CasbinAuthorize Casbin RBAC 鉴权中间件
// 通过闭包接收 *casbin.SyncedEnforcer，避免 import bootstrap 导致循环依赖
func CasbinAuthorize(enforcer *casbin.SyncedEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// enforcer 为 nil 时直接放行
		if enforcer == nil {
			c.Next()
			return
		}

		// 从 gin.Context 获取 userID
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}
		userID, ok := userIDVal.(uint)
		if !ok || userID == 0 {
			c.Next()
			return
		}

		// 从 gin.Context 获取 tenantID
		tenantIDVal, exists := c.Get("tenantID")
		if !exists {
			c.Next()
			return
		}
		tenantID, ok := tenantIDVal.(uint)
		if !ok || tenantID == 0 {
			c.Next()
			return
		}

		// 提取资源和操作
		resource := extractResource(c.Request.URL.Path)
		action := methodToAction(c.Request.Method)

		// 构造 Casbin sub：使用 "tenantID:userID" 格式
		sub := fmt.Sprintf("%d:%d", tenantID, userID)
		obj := resource

		allowed, err := enforcer.Enforce(sub, obj, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "Casbin 鉴权失败: " + err.Error(),
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

// extractResource 从请求路径提取资源名
// 移除 /api/v1/system/ 或 /api/v1/platform/ 前缀，取第一段，去掉尾部的 "s"
func extractResource(path string) string {
	p := path

	// 移除已知前缀
	prefixes := []string{"/api/v1/system/", "/api/v1/platform/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(p, prefix) {
			p = strings.TrimPrefix(p, prefix)
			break
		}
	}

	// 移除前导 /
	p = strings.TrimPrefix(p, "/")

	// 取第一段（到下一个 / 为止）
	if idx := strings.Index(p, "/"); idx > 0 {
		p = p[:idx]
	}

	// 去掉尾部的 "s"（简单英文复数转单数）
	if len(p) > 1 && strings.HasSuffix(p, "s") {
		p = p[:len(p)-1]
	}

	return p
}

// methodToAction 将 HTTP 方法映射为操作名
func methodToAction(method string) string {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return "list"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}
