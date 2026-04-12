package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	batchv1 "k8s.io/api/batch/v1"
)

// ListJobs godoc
// @Summary 获取 Job 列表
// @Description 获取 Job 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/job/list [get]
func ListJobs(c *gin.Context) {
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

	resp, err := svc.ListJobs(clusterName, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetJobDetail godoc
// @Summary 获取 Job 详情
// @Description 获取指定 Job 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/job/detail [get]
func GetJobDetail(c *gin.Context) {
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

	data, err := svc.GetJobDetail(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetJobPods godoc
// @Summary 获取 Job 关联的 Pod 列表
// @Description 获取指定 Job 关联的 Pod 列表
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/job/pods [get]
func GetJobPods(c *gin.Context) {
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

	data, err := svc.ListPodsByOwner(clusterName, namespace, "Job", name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetJobYAML godoc
// @Summary 获取 Job YAML
// @Description 获取指定 Job 的 YAML 内容
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/job/yaml [get]
func GetJobYAML(c *gin.Context) {
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

	raw, err := svc.GetJobYAML(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"yaml": raw}})
}

// CreateJob godoc
// @Summary 创建 Job
// @Description 创建一个新的 Job
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param job body object true "Job 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/job/create [post]
func CreateJob(c *gin.Context) {
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

	var job batchv1.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.CreateJob(clusterName, namespace, &job)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeleteJob godoc
// @Summary 删除 Job
// @Description 删除指定的 Job
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/job/delete [post]
func DeleteJob(c *gin.Context) {
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

	if err := svc.DeleteJob(clusterName, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
