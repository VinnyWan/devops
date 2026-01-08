package middleware

import (
	"strconv"

	"devops/common"
	k8sservice "devops/service/k8s"

	"github.com/gin-gonic/gin"
)

// K8sPermission K8s权限验证中间件
// 验证用户对指定集群的访问权限
func K8sPermission(operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID := GetUserID(c)
		if userID == 0 {
			common.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取集群ID
		clusterIDStr := c.Param("clusterId")
		if clusterIDStr == "" {
			clusterIDStr = c.Param("id")
		}

		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			common.Fail(c, "集群ID无效")
			c.Abort()
			return
		}

		// 检查权限
		permService := &k8sservice.PermissionService{}
		accessType, namespaces, err := permService.CheckAccess(userID, uint(clusterID), operation)
		if err != nil {
			common.Forbidden(c, err.Error())
			c.Abort()
			return
		}

		// 将权限信息存储到上下文
		c.Set("k8s_access_type", accessType)
		c.Set("k8s_namespaces", namespaces)
		c.Set("k8s_cluster_id", uint(clusterID))

		c.Next()
	}
}

// GetK8sAccessType 从上下文获取K8s访问类型
func GetK8sAccessType(c *gin.Context) string {
	if accessType, exists := c.Get("k8s_access_type"); exists {
		return accessType.(string)
	}
	return ""
}

// GetK8sNamespaces 从上下文获取可访问的命名空间列表
func GetK8sNamespaces(c *gin.Context) []string {
	if namespaces, exists := c.Get("k8s_namespaces"); exists {
		if ns, ok := namespaces.([]string); ok {
			return ns
		}
	}
	return nil
}

// GetK8sClusterID 从上下文获取集群ID
func GetK8sClusterID(c *gin.Context) uint {
	if clusterID, exists := c.Get("k8s_cluster_id"); exists {
		return clusterID.(uint)
	}
	return 0
}
