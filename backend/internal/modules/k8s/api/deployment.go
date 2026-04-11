package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
)

// ListDeployments godoc
// @Summary 获取 Deployment 列表
// @Description 获取 Deployment 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/deployment/list [get]
func ListDeployments(c *gin.Context) {
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

	resp, err := svc.ListDeployments(clusterName, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetDeploymentDetail godoc
// @Summary 获取 Deployment 详情
// @Description 获取指定 Deployment 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/detail [get]
func GetDeploymentDetail(c *gin.Context) {
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

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.GetDeploymentDetail(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateDeployment godoc
// @Summary 创建 Deployment
// @Description 创建一个新的 Deployment
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param deployment body K8sObject true "Deployment 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/create [post]
func CreateDeployment(c *gin.Context) {
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

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreateDeployment(clusterName, namespace, &deployment)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateDeployment godoc
// @Summary 更新 Deployment
// @Description 更新现有的 Deployment
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param deployment body K8sObject true "Deployment 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/update [post]
func UpdateDeployment(c *gin.Context) {
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

	var deployment appsv1.Deployment
	if err := c.ShouldBindJSON(&deployment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateDeployment(clusterName, namespace, &deployment)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeleteDeployment godoc
// @Summary 删除 Deployment
// @Description 删除指定的 Deployment
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name}" example({"clusterName": "k8s-prod-01", "namespace": "default", "name": "nginx-deployment"})
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/delete [post]
func DeleteDeployment(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
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

	if err := service.DeleteDeployment(clusterName, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetDeploymentPods godoc
// @Summary 根据 Deployment 获取 Pod 列表
// @Description 根据 Deployment 获取对应的 Pod 列表
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/pods [get]
func GetDeploymentPods(c *gin.Context) {
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

	data, err := svc.ListPodsByOwner(clusterName, namespace, "Deployment", name)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetDeploymentYAML godoc
// @Summary 获取 Deployment YAML
// @Description 获取指定 Deployment 的 YAML 内容
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/yaml [get]
func GetDeploymentYAML(c *gin.Context) {
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

	raw, err := svc.GetDeploymentYAML(clusterName, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"yaml": raw}})
}

// UpdateDeploymentYAML godoc
// @Summary 通过 YAML 更新 Deployment
// @Description 使用 YAML 内容更新指定 Deployment
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Param request body object true "参数: {yaml}" example({"yaml":"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx\n  namespace: default\n..."})
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/yaml/update [post]
func UpdateDeploymentYAML(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	var req struct {
		YAML string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

	updated, err := svc.UpdateDeploymentByYAML(clusterName, namespace, name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// RestartDeployment godoc
// @Summary 重启 Deployment
// @Description 通过修改 PodTemplate 注解触发滚动重启
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name}" example({"clusterName":"k8s-prod-01","namespace":"default","name":"nginx"})
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/restart [post]
func RestartDeployment(c *gin.Context) {
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

	updated, err := svc.RestartDeployment(clusterName, req.Namespace, req.Name)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// ScaleDeployment godoc
// @Summary 扩缩容 Deployment
// @Description 设置 Deployment 的副本数（支持缩容到 0）
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, replicas}" example({"clusterName":"k8s-prod-01","namespace":"default","name":"nginx","replicas":0})
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/deployment/scale [post]
func ScaleDeployment(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Replicas    *int32 `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if *req.Replicas < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "replicas 不能为负数"})
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

	updated, err := svc.ScaleDeployment(clusterName, req.Namespace, req.Name, *req.Replicas)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}
