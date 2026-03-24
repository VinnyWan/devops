package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

// ListConfigMaps godoc
// @Summary 获取 ConfigMap 列表
// @Description 获取 ConfigMap 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/configmap/list [get]
func ListConfigMaps(c *gin.Context) {
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

	resp, err := svc.ListConfigMaps(clusterID, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetConfigMapDetail godoc
// @Summary 获取 ConfigMap 详情
// @Description 获取指定 ConfigMap 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/configmap/detail [get]
func GetConfigMapDetail(c *gin.Context) {
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

	data, err := service.GetConfigMapDetail(clusterID, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateConfigMap godoc
// @Summary 创建 ConfigMap
// @Description 创建一个新的 ConfigMap
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param configmap body K8sObject true "ConfigMap 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/configmap/create [post]
func CreateConfigMap(c *gin.Context) {
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

	var cm corev1.ConfigMap
	if err := c.ShouldBindJSON(&cm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreateConfigMap(clusterID, namespace, &cm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateConfigMap godoc
// @Summary 更新 ConfigMap
// @Description 更新现有的 ConfigMap
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param configmap body K8sObject true "ConfigMap 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/configmap/update [post]
func UpdateConfigMap(c *gin.Context) {
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

	var cm corev1.ConfigMap
	if err := c.ShouldBindJSON(&cm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateConfigMap(clusterID, namespace, &cm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeleteConfigMap godoc
// @Summary 删除 ConfigMap
// @Description 删除指定的 ConfigMap
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/configmap/delete [post]
func DeleteConfigMap(c *gin.Context) {
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

	if err := service.DeleteConfigMap(clusterID, req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
