package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"

	"devops-platform/internal/modules/k8s/service"
)

// ListNodeRequest 节点列表请求参数
type ListNodeRequest struct {
	ClusterName string `form:"clusterName" binding:"required"`
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=10"`
	Name        string `form:"name"`
	Status      string `form:"status"` // Ready, NotReady
	Role        string `form:"role"`   // master, worker
}

// NodeList 获取节点列表
// @Summary 获取节点列表
// @Description 分页获取 Kubernetes 节点列表，包含 Metrics 监控数据和 Pod 数量
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param clusterName query string true "集群名称"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param name query string false "节点名称模糊搜索"
// @Param status query string false "状态过滤 (Ready, NotReady)"
// @Param role query string false "角色过滤 (master, worker)"
// @Success 200 {object} service.NodeListResponse "成功"
// @Failure 400 {object} Response "参数错误"
// @Security BearerAuth
// @Router /k8s/nodes [get]
func NodeList(c *gin.Context) {
	var req ListNodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误: " + err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败: " + err.Error()})
		return
	}

	resp, err := svc.ListNodes(req.ClusterName, req.Page, req.PageSize, req.Name, req.Status, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "获取节点列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: resp})
}

// GetNodeDetail 获取节点详情
// @Summary 获取节点详情
// @Description 获取单个节点的完整详情，包括 SystemInfo, Conditions, Images 等
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param clusterName query string true "集群名称"
// @Param name query string true "节点名称"
// @Success 200 {object} service.NodeDetail "成功"
// @Failure 400 {object} Response "参数错误"
// @Security BearerAuth
// @Router /k8s/node/detail [get]
func GetNodeDetail(c *gin.Context) {
	clusterName := c.Query("clusterName")
	name := c.Query("name")

	if clusterName == "" || name == "" {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "clusterName 和 name 不能为空"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	detail, err := svc.GetNodeDetail(clusterName, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "获取节点详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: detail})
}

// CordonNodeRequest 调度管理请求
type CordonNodeRequest struct {
	ClusterName string `json:"clusterName" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Cordon      bool   `json:"cordon"` // true=不可调度, false=可调度
}

// CordonNode 设置/取消节点调度
// @Summary 设置节点调度状态
// @Description 开启或禁用节点调度 (Cordon/Uncordon)
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param request body CordonNodeRequest true "请求参数"
// @Success 200 {object} Response "成功"
// @Failure 400 {object} Response "参数错误"
// @Security BearerAuth
// @Router /k8s/node/cordon [post]
func CordonNode(c *gin.Context) {
	var req CordonNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	if err := svc.CordonNode(req.ClusterName, req.Name, req.Cordon); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "设置调度状态失败: " + err.Error()})
		return
	}

	action := "恢复调度"
	if req.Cordon {
		action = "停止调度"
	}
	c.JSON(http.StatusOK, Response{Code: 200, Message: action + "成功"})
}

// DrainNodeRequest 节点驱逐请求
type DrainNodeRequest struct {
	ClusterName        string `json:"clusterName" binding:"required"`
	Name               string `json:"name" binding:"required"`
	GracePeriodSeconds int    `json:"gracePeriodSeconds"`
	Force              bool   `json:"force"`
	IgnoreDaemonSets   bool   `json:"ignoreDaemonSets"`
	DeleteLocalData    bool   `json:"deleteLocalData"`
}

// DrainNode 驱逐节点
// @Summary 驱逐节点 (Drain)
// @Description 安全驱逐节点上的 Pod，支持 Force 和 IgnoreDaemonSets 选项
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param request body DrainNodeRequest true "请求参数"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/node/drain [post]
func DrainNode(c *gin.Context) {
	var req DrainNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	opts := service.DrainOptions{
		GracePeriodSeconds: req.GracePeriodSeconds,
		Force:              req.Force,
		IgnoreDaemonSets:   req.IgnoreDaemonSets,
		DeleteLocalData:    req.DeleteLocalData,
	}

	if err := svc.DrainNode(req.ClusterName, req.Name, opts); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "节点驱逐失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "节点驱逐成功"})
}

// UpdateLabelsRequest 更新标签请求
type UpdateLabelsRequest struct {
	ClusterName string            `json:"clusterName" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Labels      map[string]string `json:"labels" binding:"required"`
}

// UpdateNodeLabels 更新节点标签
// @Summary 更新节点标签
// @Description 批量更新节点标签 (全量替换)
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param request body UpdateLabelsRequest true "请求参数"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/node/labels [post]
func UpdateNodeLabels(c *gin.Context) {
	var req UpdateLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	// 使用 Revised 版本 (Retry Update)
	if err := svc.UpdateNodeLabels_Revised(req.ClusterName, req.Name, req.Labels); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "更新标签失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "更新标签成功"})
}

// UpdateTaintsRequest 更新污点请求
type UpdateTaintsRequest struct {
	ClusterName string        `json:"clusterName" binding:"required"`
	Name        string        `json:"name" binding:"required"`
	Taints      []interface{} `json:"taints" binding:"required"`
}

// UpdateNodeTaints 更新节点污点
// @Summary 更新节点污点
// @Description 批量更新节点污点 (全量替换列表)
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param request body UpdateTaintsRequest true "请求参数"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/node/taints [post]
func UpdateNodeTaints(c *gin.Context) {
	var req UpdateTaintsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "参数错误"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	// 使用 Revised 版本 (Retry Update)
	var taints []corev1.Taint
	taintsBytes, _ := json.Marshal(req.Taints)
	json.Unmarshal(taintsBytes, &taints)

	if err := svc.UpdateNodeTaints_Revised(req.ClusterName, req.Name, taints); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "更新污点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "更新污点成功"})
}

// GetNodeEvents 获取节点事件
// @Summary 获取节点事件
// @Description 获取与指定节点相关的 K8s 事件
// @Tags K8s节点管理
// @Accept json
// @Produce json
// @Param clusterName query string true "集群名称"
// @Param name query string true "节点名称"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/node/events [get]
func GetNodeEvents(c *gin.Context) {
	clusterName := c.Query("clusterName")
	name := c.Query("name")

	if clusterName == "" || name == "" {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Message: "clusterName 和 name 不能为空"})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "服务初始化失败"})
		return
	}

	events, err := svc.GetNodeEvents(clusterName, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 500, Message: "获取事件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 200, Message: "获取成功", Data: events})
}
