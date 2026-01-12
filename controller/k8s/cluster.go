package k8s

import (
	"strconv"

	"devops/common"
	k8smodels "devops/models/k8s"
	k8sservice "devops/service/k8s"

	"github.com/gin-gonic/gin"
)

type ClusterController struct {
	clusterService    *k8sservice.ClusterService
	permissionService *k8sservice.PermissionService
}

func NewClusterController() *ClusterController {
	return &ClusterController{
		clusterService:    &k8sservice.ClusterService{},
		permissionService: &k8sservice.PermissionService{},
	}
}

// Create 创建集群
// @Summary 创建K8s集群
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param data body object true "集群信息"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters [post]
func (ctrl *ClusterController) Create(c *gin.Context) {
	var cluster k8smodels.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.clusterService.Create(&cluster); err != nil {
		common.Fail(c, "创建集群失败: "+err.Error())
		return
	}

	common.Success(c, cluster)
}

// GetList 获取集群列表
// @Summary 获取K8s集群列表
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param name query string false "集群名称"
// @Param deptId query int false "部门ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters [get]
func (ctrl *ClusterController) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")
	deptID, _ := strconv.ParseUint(c.Query("deptId"), 10, 32)

	// 获取当前用户ID（注意：JWT中间件使用的key是 "userId"，不是 "userID"）
	userID, exists := c.Get("userId")
	if !exists {
		common.Unauthorized(c, "未登录")
		return
	}

	// 将 userId 转换为 uint 类型
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	default:
		common.Unauthorized(c, "用户ID格式错误")
		return
	}

	clusters, total, err := ctrl.clusterService.GetListByUser(uid, page, pageSize, name, uint(deptID))
	if err != nil {
		common.Fail(c, "获取集群列表失败")
		return
	}

	common.PageSuccess(c, clusters, total, page, pageSize)
}

// GetByID 根据ID获取集群
// @Summary 获取K8s集群详情
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId} [get]
func (ctrl *ClusterController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	cluster, err := ctrl.clusterService.GetByID(uint(id))
	if err != nil {
		common.Fail(c, "集群不存在")
		return
	}

	common.Success(c, cluster)
}

// Update 更新集群
// @Summary 更新K8s集群
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Param data body object true "集群信息"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId} [put]
func (ctrl *ClusterController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	var cluster k8smodels.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		common.Fail(c, "参数错误")
		return
	}

	if err := ctrl.clusterService.Update(uint(id), &cluster); err != nil {
		common.Fail(c, "更新集群失败: "+err.Error())
		return
	}

	common.Success(c, nil)
}

// Delete 删除集群
// @Summary 删除K8s集群
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId} [delete]
func (ctrl *ClusterController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err := ctrl.clusterService.Delete(uint(id)); err != nil {
		common.Fail(c, "删除集群失败")
		return
	}

	common.Success(c, nil)
}

// HealthCheck 健康检查
// @Summary K8s集群健康检查
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId}/health [get]
func (ctrl *ClusterController) HealthCheck(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	healthy, version, err := ctrl.clusterService.HealthCheck(uint(id))

	if err != nil {
		common.Success(c, gin.H{
			"healthy": false,
			"version": "",
			"message": err.Error(),
		})
		return
	}

	common.Success(c, gin.H{
		"healthy": healthy,
		"version": version,
		"message": "集群连接正常",
	})
}

// ReimportKubeConfig 重新导入KubeConfig
// @Summary 重新导入K8s集群配置
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Param data body object true "KubeConfig配置"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId}/reimport [post]
func (ctrl *ClusterController) ReimportKubeConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

	var req struct {
		KubeConfig string `json:"kubeConfig" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.clusterService.ReimportKubeConfig(uint(id), req.KubeConfig); err != nil {
		common.Fail(c, "重新导入失败: "+err.Error())
		return
	}

	// 获取更新后的集群信息
	cluster, err := ctrl.clusterService.GetByID(uint(id))
	if err != nil {
		common.Success(c, gin.H{
			"message": "重新导入成功",
		})
		return
	}

	common.Success(c, cluster)
}

// CreateAccess 创建集群访问权限
// @Summary 创建集群访问权限
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Param data body object true "权限信息"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId}/access [post]
func (ctrl *ClusterController) CreateAccess(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

	var access k8smodels.ClusterAccess
	if err := c.ShouldBindJSON(&access); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}

	access.ClusterID = uint(clusterID)

	if err := ctrl.permissionService.CreateAccess(&access); err != nil {
		common.Fail(c, "创建权限失败: "+err.Error())
		return
	}

	common.Success(c, access)
}

// GetAccessList 获取集群访问权限列表
// @Summary 获取集群访问权限列表
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId}/access [get]
func (ctrl *ClusterController) GetAccessList(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("clusterId"), 10, 32)

	accesses, err := ctrl.permissionService.GetAccessList(uint(clusterID))
	if err != nil {
		common.Fail(c, "获取权限列表失败")
		return
	}

	common.Success(c, accesses)
}

// DeleteAccess 删除集群访问权限
// @Summary 删除集群访问权限
// @Tags K8s-Cluster
// @Accept json
// @Produce json
// @Param clusterId path int true "集群ID"
// @Param accessId path int true "权限ID"
// @Success 200 {object} common.Response
// @Security Bearer
// @Router /api/k8s/clusters/{clusterId}/access/{accessId} [delete]
func (ctrl *ClusterController) DeleteAccess(c *gin.Context) {
	accessID, _ := strconv.ParseUint(c.Param("accessId"), 10, 32)

	if err := ctrl.permissionService.DeleteAccess(uint(accessID)); err != nil {
		common.Fail(c, "删除权限失败")
		return
	}

	common.Success(c, nil)
}
