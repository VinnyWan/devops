package api

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ValidateK8sParams 验证K8s资源操作的通用参数
func ValidateK8sParams(c *gin.Context) (clusterName, namespace, name string, err error) {
	clusterName = c.Query("clusterName")
	if clusterName == "" {
		err = errors.New("clusterName is required")
		return
	}
	namespace = c.Query("namespace")
	if namespace == "" {
		err = errors.New("namespace is required")
		return
	}
	name = c.Query("name")
	if name == "" {
		err = errors.New("name is required")
		return
	}
	return
}

// ResolveClusterName 解析集群名称
func ResolveClusterName(c *gin.Context) (string, error) {
	clusterName := c.Query("clusterName")
	if clusterName == "" {
		return "", errors.New("clusterName is required")
	}
	return clusterName, nil
}
