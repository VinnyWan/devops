package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

// ListPods godoc
// @Summary 获取 Pod 列表
// @Description 获取 Pod 列表，namespace 为空时获取所有命名空间
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string false "命名空间（为空时查询所有命名空间）"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param keyword query string false "关键字搜索（匹配名称）"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/list [get]
func ListPods(c *gin.Context) {
	var req K8sListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误: " + err.Error()})
		return
	}

	clusterID, err := resolveListClusterID(c, req.ClusterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	resp, err := svc.ListPods(clusterID, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetPodDetail godoc
// @Summary 获取 Pod 详情
// @Description 获取指定 Pod 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/detail [get]
func GetPodDetail(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.GetPodDetail(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreatePod godoc
// @Summary 创建 Pod
// @Description 创建一个新的 Pod
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param pod body K8sObject true "Pod 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/create [post]
func CreatePod(c *gin.Context) {
	namespace := c.Query("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var pod corev1.Pod
	if err := c.ShouldBindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreatePod(clusterID, namespace, &pod)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdatePod godoc
// @Summary 更新 Pod
// @Description 更新现有的 Pod
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param pod body K8sObject true "Pod 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/update [post]
func UpdatePod(c *gin.Context) {
	namespace := c.Query("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var pod corev1.Pod
	if err := c.ShouldBindJSON(&pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdatePod(clusterID, namespace, &pod)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeletePod godoc
// @Summary 删除 Pod
// @Description 删除指定的 Pod
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/delete [post]
func DeletePod(c *gin.Context) {
	var req struct {
		ClusterID uint   `json:"clusterId"`
		Namespace string `json:"namespace" binding:"required"`
		Name      string `json:"name" binding:"required"`
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

	clusterID := req.ClusterID
	if clusterID == 0 {
		clusterID, err = resolveClusterID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if err := service.DeletePod(clusterID, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ListPodsByOwner godoc
// @Summary 根据控制器获取 Pod 列表
// @Description 根据 Deployment/StatefulSet/DaemonSet 获取 Pod 列表
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param ownerType query string true "控制器类型 (Deployment/StatefulSet/DaemonSet)"
// @Param ownerName query string true "控制器名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/list_by_owner [post]
// DescribePod godoc
// @Summary 获取 Pod 诊断信息
// @Description 聚合 Pod 基础信息、容器状态、事件、卷等完整诊断数据
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "Pod 名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/describe [get]
func DescribePod(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := svc.DescribePod(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func ListPodsByOwner(c *gin.Context) {
	namespace := c.Query("namespace")
	ownerType := c.Query("ownerType")
	ownerName := c.Query("ownerName")

	if namespace == "" || ownerType == "" || ownerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.ListPodsByOwner(clusterID, namespace, ownerType, ownerName)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}
