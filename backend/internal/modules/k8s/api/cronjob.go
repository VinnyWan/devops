package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
)

// ListCronJobs godoc
// @Summary 获取 CronJob 列表
// @Description 获取 CronJob 列表，namespace 为空时获取所有命名空间
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string false "命名空间（为空时查询所有命名空间）"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "关键字搜索（匹配名称）"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/list [get]
func ListCronJobs(c *gin.Context) {
	var req K8sListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误: " + err.Error()})
		return
	}

	clusterName, err := resolveListClusterName(c, req.ClusterName)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	resp, err := svc.ListCronJobs(clusterName, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetCronJobDetail godoc
// @Summary 获取 CronJob 详情
// @Description 获取指定 CronJob 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/detail [get]
func GetCronJobDetail(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterName, err := resolveClusterName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.GetCronJobDetail(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetCronJobPods godoc
// @Summary 获取 CronJob 关联的 Pod 列表
// @Description 获取指定 CronJob 关联的 Pod 列表
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/pods [get]
func GetCronJobPods(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterName, err := resolveClusterName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.ListPodsByOwner(clusterName, namespace, "CronJob", name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetCronJobYAML godoc
// @Summary 获取 CronJob YAML
// @Description 获取指定 CronJob 的 YAML 内容
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/yaml [get]
func GetCronJobYAML(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterName, err := resolveClusterName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	raw, err := svc.GetCronJobYAML(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"yaml": raw}})
}

// CreateCronJob godoc
// @Summary 创建 CronJob
// @Description 创建一个新的 CronJob
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param cronjob body object true "CronJob 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/create [post]
func CreateCronJob(c *gin.Context) {
	namespace := c.Query("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterName, err := resolveClusterName(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var cj batchv1.CronJob
	if err := c.ShouldBindJSON(&cj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.CreateCronJob(clusterName, namespace, &cj)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateCronJobYAML godoc
// @Summary 通过 YAML 更新 CronJob
// @Description 使用 YAML 内容更新指定 CronJob
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, yaml}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/yaml/update [post]
func UpdateCronJobYAML(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		YAML        string `json:"yaml" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		var err error
		clusterName, err = resolveClusterName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	updated, err := svc.UpdateCronJobByYAML(clusterName, req.Namespace, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// SuspendCronJob godoc
// @Summary 暂停/恢复 CronJob
// @Description 暂停或恢复指定的 CronJob
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, suspend}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/suspend [post]
func SuspendCronJob(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Suspend     bool   `json:"suspend"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		var err error
		clusterName, err = resolveClusterName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	updated, err := svc.SuspendCronJob(clusterName, req.Namespace, req.Name, req.Suspend)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// DeleteCronJob godoc
// @Summary 删除 CronJob
// @Description 删除指定的 CronJob
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/cronjob/delete [post]
func DeleteCronJob(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	clusterName := req.ClusterName
	if clusterName == "" {
		clusterName, err = resolveClusterName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if err := svc.DeleteCronJob(clusterName, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
