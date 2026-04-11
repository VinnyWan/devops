package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

// ListServices godoc
// @Summary 获取 Service 列表
// @Description 获取 Service 列表，namespace 为空时获取所有命名空间
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
// @Router /k8s/service/list [get]
func ListServices(c *gin.Context) {
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

	resp, err := svc.ListServices(clusterName, req.Namespace, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetServiceDetail godoc
// @Summary 获取 Service详情
// @Description 获取指定 Service 的详细信息
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/service/detail [get]
func GetServiceDetail(c *gin.Context) {
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

	data, err := service.GetServiceDetail(clusterName, namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// CreateService godoc
// @Summary 创建 Service
// @Description 创建一个新的 Service
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param service body K8sObject true "Service 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/service/create [post]
func CreateService(c *gin.Context) {
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

	var svc corev1.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.CreateService(clusterName, namespace, &svc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdateService godoc
// @Summary 更新 Service
// @Description 更新现有的 Service
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param service body K8sObject true "Service 对象"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/service/update [post]
func UpdateService(c *gin.Context) {
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

	var svc corev1.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	service, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	data, err := service.UpdateService(clusterName, namespace, &svc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeleteService godoc
// @Summary 删除 Service
// @Description 删除指定的 Service
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "参数: {clusterName, namespace, name}"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/service/delete [post]
func DeleteService(c *gin.Context) {
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

	if err := service.DeleteService(clusterName, req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
