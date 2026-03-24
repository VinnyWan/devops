package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NamespaceList godoc
// @Summary 获取 Namespace 列表
// @Description 获取指定集群下的所有 Namespace 列表
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Success 200 {object} Response "成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Security BearerAuth
// @Router /k8s/namespaces/list [get]
func NamespaceList(c *gin.Context) {
	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务初始化失败",
			"error":   err.Error(),
		})
		return
	}

	data, err := service.ListNamespaces(clusterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取 Namespace 失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    data,
	})
}
