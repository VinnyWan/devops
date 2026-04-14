package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListStorageClasses godoc
// @Summary 获取 StorageClass 列表
// @Tags K8s资源管理
// @Produce json
// @Param clusterName query string false "集群名称"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/storageclass/list [get]
func ListStorageClasses(c *gin.Context) {
	clusterName, err := resolveListClusterName(c, c.Query("clusterName"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	data, err := svc.ListStorageClasses(clusterName)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: data})
}

// ListPersistentVolumes godoc
// @Summary 获取 PV 列表
// @Tags K8s资源管理
// @Produce json
// @Param clusterName query string false "集群名称"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/pv/list [get]
func ListPersistentVolumes(c *gin.Context) {
	clusterName, err := resolveListClusterName(c, c.Query("clusterName"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	data, err := svc.ListPersistentVolumes(clusterName)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: data})
}

// ListPersistentVolumeClaims godoc
// @Summary 获取 PVC 列表
// @Tags K8s资源管理
// @Produce json
// @Param clusterName query string false "集群名称"
// @Param namespace query string false "命名空间"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/pvc/list [get]
func ListPersistentVolumeClaims(c *gin.Context) {
	clusterName, err := resolveListClusterName(c, c.Query("clusterName"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	data, err := svc.ListPersistentVolumeClaims(clusterName, c.Query("namespace"))
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: data})
}

// UpdateStorageClassYAML godoc
// @Summary 通过 YAML 更新 StorageClass
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, name, yaml}"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/storageclass/yaml/update [post]
func UpdateStorageClassYAML(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
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

	data, err := svc.UpdateStorageClassByYAML(clusterName, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdatePVYAML godoc
// @Summary 通过 YAML 更新 PV
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, name, yaml}"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/pv/yaml/update [post]
func UpdatePVYAML(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
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

	data, err := svc.UpdatePVByYAML(clusterName, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// UpdatePVCYAML godoc
// @Summary 通过 YAML 更新 PVC
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name, yaml}"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/pvc/yaml/update [post]
func UpdatePVCYAML(c *gin.Context) {
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

	data, err := svc.UpdatePVCByYAML(clusterName, req.Namespace, req.Name, req.YAML)
	if err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

// DeletePVC godoc
// @Summary 删除 PVC
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param request body object true "参数: {clusterName, namespace, name}"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /k8s/pvc/delete [post]
func DeletePVC(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
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

	if err := svc.DeletePVC(clusterName, req.Namespace, req.Name); err != nil {
		handleK8sError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
