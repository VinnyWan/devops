package api

import (
	"net/http"

	"devops-platform/internal/modules/k8s/service"

	"github.com/gin-gonic/gin"
)

// GetResourceYAMLRequest 获取资源YAML请求
type GetResourceYAMLRequest struct {
	ClusterName  string `form:"clusterName" json:"clusterName"`
	ResourceType string `form:"resourceType" json:"resourceType" binding:"required"` // pod, deployment, service 等
	Namespace    string `form:"namespace" json:"namespace"`                          // 命名空间
	Name         string `form:"name" json:"name" binding:"required"`                 // 资源名称
}

// GetResourceYAML godoc
// @Summary 获取资源YAML
// @Description 通用接口，获取任意K8s资源的YAML格式
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterName query string false "集群名称（可选，未传则使用默认集群）"
// @Param resourceType query string true "资源类型（pod/deployment/service/ingress等）"
// @Param namespace query string false "命名空间（集群级别资源可不传）"
// @Param name query string true "资源名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/resource/yaml [get]
func GetResourceYAML(c *gin.Context) {
	var req GetResourceYAMLRequest
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

	yaml, err := svc.GetResourceYAML(clusterName, req.ResourceType, req.Namespace, req.Name)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: gin.H{
		"yaml": yaml,
	}})
}

// GetSupportedResourceTypes godoc
// @Summary 获取支持的资源类型
// @Description 返回通用YAML接口支持的资源类型列表
// @Tags K8s资源管理
// @Produce json
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/resource/types [get]
func GetSupportedResourceTypes(c *gin.Context) {
	types := service.GetSupportedResourceTypes()
	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: gin.H{
		"types": types,
	}})
}
