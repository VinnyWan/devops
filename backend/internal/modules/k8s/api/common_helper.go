package api

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidateK8sParams 验证K8s资源操作的通用参数
func ValidateK8sParams(c *gin.Context) (clusterID uint, namespace, name string, err error) {
	clusterIDStr := c.Query("cluster_id")
	if clusterIDStr == "" {
		err = errors.New("cluster_id is required")
		return
	}

	id, parseErr := strconv.ParseUint(clusterIDStr, 10, 32)
	if parseErr != nil {
		err = errors.New("invalid cluster_id")
		return
	}
	clusterID = uint(id)

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

// ResolveClusterID 解析集群ID
func ResolveClusterID(c *gin.Context) (uint, error) {
	clusterIDStr := c.Query("cluster_id")
	if clusterIDStr == "" {
		return 0, errors.New("cluster_id is required")
	}

	id, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		return 0, errors.New("invalid cluster_id")
	}

	return uint(id), nil
}
