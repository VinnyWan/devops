package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ClusterWorkloadStats 获取工作负载统计
// @Summary 获取工作负载统计
// @Description 统计 Deployment, StatefulSet, DaemonSet, Job, CronJob 的数量
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Success 200 {object} Response "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Security BearerAuth
// @Router /k8s/cluster/stats/workload [get]
func ClusterWorkloadStats(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id 不能为空"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的集群ID"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetWorkloadCounts(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取统计失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    counts,
	})
}

// ClusterNetworkStats 获取网络资源统计
// @Summary 获取网络资源统计
// @Description 统计 Service, Ingress 的数量
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/stats/network [get]
func ClusterNetworkStats(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id 不能为空"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的集群ID"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetNetworkCounts(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取统计失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    counts,
	})
}

// ClusterStorageStats 获取存储资源统计
// @Summary 获取存储资源统计
// @Description 统计 PV, PVC 的数量
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/stats/storage [get]
func ClusterStorageStats(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id 不能为空"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的集群ID"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetStorageCounts(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取统计失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    counts,
	})
}

// ClusterNodes 获取节点列表
// @Summary 获取节点列表
// @Description 获取节点列表，包含资源使用情况（若有 Metrics Server）
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param name query string false "节点名称搜索"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/nodes [get]
func ClusterNodes(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id 不能为空"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的集群ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	result, err := svc.GetNodeList(uint(id), page, pageSize, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取节点列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    result,
	})
}

// ClusterEvents 获取事件列表
// @Summary 获取事件列表
// @Description 获取集群事件列表
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/events [get]
func ClusterEvents(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "id 不能为空"})
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的集群ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	result, err := svc.GetEventList(uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取事件列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    result,
	})
}
