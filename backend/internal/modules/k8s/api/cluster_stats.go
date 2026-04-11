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
// @Param name query string true "集群名称"
// @Success 200 {object} Response "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Security BearerAuth
// @Router /k8s/cluster/stats/workload [get]
func ClusterWorkloadStats(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "name 不能为空"})
		return
	}

	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	clusterSvc := getService()
	cluster, err := clusterSvc.GetByExactNameInTenant(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "集群不存在", "error": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetWorkloadCounts(cluster.Name)
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
// @Param name query string true "集群名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/stats/network [get]
func ClusterNetworkStats(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "name 不能为空"})
		return
	}

	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	clusterSvc := getService()
	cluster, err := clusterSvc.GetByExactNameInTenant(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "集群不存在", "error": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetNetworkCounts(cluster.Name)
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
// @Param name query string true "集群名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/stats/storage [get]
func ClusterStorageStats(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "name 不能为空"})
		return
	}

	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	clusterSvc := getService()
	cluster, err := clusterSvc.GetByExactNameInTenant(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "集群不存在", "error": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	counts, err := svc.GetStorageCounts(cluster.Name)
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
// @Param name query string true "集群名称"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param name query string false "节点名称搜索"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/nodes [get]
func ClusterNodes(c *gin.Context) {
	clusterName := c.Query("name")
	if clusterName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "name 不能为空"})
		return
	}

	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	clusterSvc := getService()
	cluster, err := clusterSvc.GetByExactNameInTenant(tenantID, clusterName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "集群不存在", "error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	nodeName := c.Query("nodeName")

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	result, err := svc.GetNodeList(cluster.Name, page, pageSize, nodeName)
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
// @Param name query string true "集群名称"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cluster/events [get]
func ClusterEvents(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "name 不能为空"})
		return
	}

	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	clusterSvc := getService()
	cluster, err := clusterSvc.GetByExactNameInTenant(tenantID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "集群不存在", "error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化", "error": err.Error()})
		return
	}

	result, err := svc.GetEventList(cluster.Name, page, pageSize)
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
