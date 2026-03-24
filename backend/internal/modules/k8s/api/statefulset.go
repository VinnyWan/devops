package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
)

// ListStatefulSets godoc
// @Summary 获取 StatefulSet 列表
// @Description 获取 StatefulSet 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/statefulset/list [get]
func ListStatefulSets(c *gin.Context) {
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

	resp, err := svc.ListStatefulSets(clusterID, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetStatefulSetDetail godoc
// @Summary 获取 StatefulSet 详情
// @Description 获取指定 StatefulSet 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/detail [get]
func GetStatefulSetDetail(c *gin.Context) {
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

	data, err := service.GetStatefulSetDetail(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// GetStatefulSetYAML godoc
// @Summary 获取 StatefulSet YAML
// @Description 获取指定 StatefulSet 的 YAML 内容
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/yaml [get]
func GetStatefulSetYAML(c *gin.Context) {
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

	raw, err := svc.GetStatefulSetYAML(clusterID, namespace, name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"yaml": raw}})
}

// UpdateStatefulSetYAML godoc
// @Summary 通过 YAML 更新 StatefulSet
// @Description 使用 YAML 内容更新指定 StatefulSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Param request body object true "参数: {yaml}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/yaml/update [post]
func UpdateStatefulSetYAML(c *gin.Context) {
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

	updated, err := svc.UpdateStatefulSetByYAML(clusterID, namespace, name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// RestartStatefulSet godoc
// @Summary 重启 StatefulSet
// @Description 通过修改 PodTemplate 注解触发滚动重启
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/restart [post]
func RestartStatefulSet(c *gin.Context) {
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

	updated, err := svc.RestartStatefulSet(clusterID, req.Namespace, req.Name)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// ScaleStatefulSet godoc
// @Summary 扩缩容 StatefulSet
// @Description 设置 StatefulSet 的副本数（支持缩容到 0）
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterId, namespace, name, replicas}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/scale [post]
func ScaleStatefulSet(c *gin.Context) {
	var req struct {
		ClusterID uint   `json:"clusterId"`
		Namespace string `json:"namespace" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Replicas  *int32 `json:"replicas" binding:"required"`
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

	clusterID := req.ClusterID
	if clusterID == 0 {
		clusterID, err = resolveClusterID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	updated, err := svc.ScaleStatefulSet(clusterID, req.Namespace, req.Name, *req.Replicas)
	if err != nil {
		handleK8sError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": updated})
}

// DeleteStatefulSet godoc
// @Summary 删除 StatefulSet
// @Description 删除指定的 StatefulSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterId, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/delete [post]
func DeleteStatefulSet(c *gin.Context) {
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

	if err := service.DeleteStatefulSet(clusterID, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// CreateStatefulSet godoc
// @Summary 创建 StatefulSet
// @Description 创建一个新的 StatefulSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param statefulset body K8sObject true "StatefulSet 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/create [post]
func CreateStatefulSet(c *gin.Context) {
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

	var statefulset appsv1.StatefulSet
	if err := c.ShouldBindJSON(&statefulset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreateStatefulSet(clusterID, namespace, &statefulset)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateStatefulSet godoc
// @Summary 更新 StatefulSet
// @Description 更新现有的 StatefulSet
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param statefulset body K8sObject true "StatefulSet 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/statefulset/update [post]
func UpdateStatefulSet(c *gin.Context) {
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

	var statefulset appsv1.StatefulSet
	if err := c.ShouldBindJSON(&statefulset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateStatefulSet(clusterID, namespace, &statefulset)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}
