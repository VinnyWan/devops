package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
)

// ListDaemonSets godoc
// @Summary 获取 DaemonSet 列表
// @Description 获取 DaemonSet 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/daemonset/list [get]
func ListDaemonSets(c *gin.Context) {
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

	resp, err := svc.ListDaemonSets(clusterID, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetDaemonSetDetail godoc
// @Summary 获取 DaemonSet 详情
// @Description 获取指定 DaemonSet 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/detail [get]
func GetDaemonSetDetail(c *gin.Context) {
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

	data, err := service.GetDaemonSetDetail(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetDaemonSetYAML godoc
// @Summary 获取 DaemonSet YAML
// @Description 获取指定 DaemonSet 的 YAML 内容
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/yaml [get]
func GetDaemonSetYAML(c *gin.Context) {
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

	raw, err := svc.GetDaemonSetYAML(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"yaml": raw}})
}

// UpdateDaemonSetYAML godoc
// @Summary 通过 YAML 更新 DaemonSet
// @Description 使用 YAML 内容更新指定 DaemonSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Param request body object true "参数: {yaml}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/yaml/update [post]
func UpdateDaemonSetYAML(c *gin.Context) {
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

	updated, err := svc.UpdateDaemonSetByYAML(clusterID, namespace, name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// RestartDaemonSet godoc
// @Summary 重启 DaemonSet
// @Description 通过修改 PodTemplate 注解触发滚动重启
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/restart [post]
func RestartDaemonSet(c *gin.Context) {
	var req struct {
		ClusterID uint   `json:"clusterId"`
		Namespace string `json:"namespace" binding:"required"`
		Name      string `json:"name" binding:"required"`
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

	clusterID := req.ClusterID
	if clusterID == 0 {
		clusterID, err = resolveClusterID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	updated, err := svc.RestartDaemonSet(clusterID, req.Namespace, req.Name)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// DeleteDaemonSet godoc
// @Summary 删除 DaemonSet
// @Description 删除指定的 DaemonSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/delete [post]
func DeleteDaemonSet(c *gin.Context) {
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

	if err := service.DeleteDaemonSet(clusterID, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// CreateDaemonSet godoc
// @Summary 创建 DaemonSet
// @Description 创建一个新的 DaemonSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param daemonset body K8sObject true "DaemonSet 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/create [post]
func CreateDaemonSet(c *gin.Context) {
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

	var daemonset appsv1.DaemonSet
	if err := c.ShouldBindJSON(&daemonset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreateDaemonSet(clusterID, namespace, &daemonset)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateDaemonSet godoc
// @Summary 更新 DaemonSet
// @Description 更新现有的 DaemonSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param daemonset body K8sObject true "DaemonSet 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/daemonset/update [post]
func UpdateDaemonSet(c *gin.Context) {
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

	var daemonset appsv1.DaemonSet
	if err := c.ShouldBindJSON(&daemonset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateDaemonSet(clusterID, namespace, &daemonset)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}
