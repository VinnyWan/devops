package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"devops-platform/internal/modules/k8s/service"
)

// getService 获取服务实例（延迟初始化）
func getService() *service.ClusterService {
	clusterOnce.Do(func() {
		clusterServiceInstance = service.NewClusterService(k8sDB)
	})
	return clusterServiceInstance
}

// ClusterResponse 集群响应
type ClusterResponse struct {
	ID         uint   `json:"id" example:"1"`
	Name       string `json:"name" example:"k8s-prod-01"`
	Url        string `json:"url" example:"https://1.2.3.4:6443"`
	AuthType   string `json:"authType" example:"kubeconfig"`
	Status     string `json:"status" example:"healthy"`
	IsDefault  bool   `json:"isDefault" example:"false"`
	Remark     string `json:"remark" example:"备注信息"`
	Labels     string `json:"labels" example:"{\"env\":\"prod\"}"`
	Env        string `json:"env" example:"prod"`
	K8sVersion string `json:"k8sVersion" example:"v1.24.0"`
	NodeCount  int    `json:"nodeCount" example:"3"`
	CreatedAt  string `json:"createdAt" example:"2023-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updatedAt" example:"2023-01-01T00:00:00Z"`
}

// ClusterListResponse 集群列表响应
type ClusterListResponse struct {
	Message  string            `json:"message" example:"获取成功"`
	Data     []ClusterResponse `json:"data"`
	Total    int64             `json:"total" example:"100"`
	Page     int               `json:"page" example:"1"`
	PageSize int               `json:"pageSize" example:"10"`
}

// List 获取集群列表
// @Summary 获取集群列表
// @Description 获取所有注册的Kubernetes集群列表，支持分页和多维度筛选
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param env query string false "环境筛选" Enums(dev, test, prod)
// @Param keyword query string false "统一搜索关键词（长度>=3生效）"
// @Param name query string false "集群名称（模糊查询）"
// @Success 200 {object} ClusterListResponse "成功"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Security BearerAuth
// @Router /k8s/cluster/list [get]
func ClusterList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	env := c.Query("env")
	name := c.Query("name")
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = name
	}

	clusters, total, err := getService().List(page, pageSize, env, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取集群列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 隐藏敏感信息
	var result []map[string]interface{}
	for _, cluster := range clusters {
		result = append(result, map[string]interface{}{
			"id":         cluster.ID,
			"name":       cluster.Name,
			"url":        cluster.Url,
			"authType":   cluster.AuthType,
			"status":     cluster.Status,
			"isDefault":  cluster.IsDefault,
			"remark":     cluster.Remark,
			"labels":     cluster.Labels,
			"env":        cluster.Env,
			"k8sVersion": cluster.K8sVersion,
			"nodeCount":  cluster.NodeCount,
			"createdAt":  cluster.CreatedAt,
			"updatedAt":  cluster.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// Detail 获取集群详情
// @Summary 获取集群详情
// @Description 根据集群ID获取详细信息
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param id query int true "集群ID"
// @Success 200 {object} ClusterResponse "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "集群不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Security BearerAuth
// @Router /k8s/cluster/detail [get]
func ClusterDetail(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "id 不能为空",
		})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的集群ID",
		})
		return
	}

	cluster, err := getService().GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "集群不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取集群详情失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"id":         cluster.ID,
			"name":       cluster.Name,
			"url":        cluster.Url,
			"authType":   cluster.AuthType,
			"status":     cluster.Status,
			"isDefault":  cluster.IsDefault,
			"remark":     cluster.Remark,
			"labels":     cluster.Labels,
			"env":        cluster.Env,
			"k8sVersion": cluster.K8sVersion,
			"nodeCount":  cluster.NodeCount,
			"createdAt":  cluster.CreatedAt,
			"updatedAt":  cluster.UpdatedAt,
		},
	})
}

// ClusterDefault godoc
// @Summary 获取默认集群
// @Description 获取当前默认集群；若未设置默认集群，则返回最新创建的一个集群
// @Tags 集群管理
// @Accept json
// @Produce json
// @Success 200 {object} ClusterResponse "成功"
// @Failure 404 {object} map[string]interface{} "无可用集群"
// @Security BearerAuth
// @Router /k8s/cluster/default [get]
func ClusterDefault(c *gin.Context) {
	cluster, err := getService().GetDefaultOrFirst()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "无可用集群",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取默认集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"id":         cluster.ID,
			"name":       cluster.Name,
			"url":        cluster.Url,
			"authType":   cluster.AuthType,
			"status":     cluster.Status,
			"isDefault":  cluster.IsDefault,
			"remark":     cluster.Remark,
			"labels":     cluster.Labels,
			"env":        cluster.Env,
			"k8sVersion": cluster.K8sVersion,
			"nodeCount":  cluster.NodeCount,
			"createdAt":  cluster.CreatedAt,
			"updatedAt":  cluster.UpdatedAt,
		},
	})
}

// ClusterSetDefault godoc
// @Summary 设置默认集群
// @Description 设置指定集群为默认集群
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "参数 {id: 1}"
// @Success 200 {object} map[string]interface{} "设置成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "集群不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Security BearerAuth
// @Router /k8s/cluster/set-default [post]
func ClusterSetDefault(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := getService().SetDefault(req.ID); err != nil {
		if err.Error() == "集群不存在" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "设置默认集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设置成功",
	})
}

// ClusterCreateResponse 集群创建/更新响应
type ClusterCreateResponse struct {
	Message string `json:"message" example:"添加集群成功"`
	Data    struct {
		ID       uint   `json:"id" example:"1"`
		Name     string `json:"name" example:"k8s-prod-01"`
		Url      string `json:"url" example:"https://1.2.3.4:6443"`
		AuthType string `json:"authType" example:"kubeconfig"`
		Status   string `json:"status" example:"healthy"`
	} `json:"data"`
}

// ClusterDeleteResponse 集群删除响应
type ClusterDeleteResponse struct {
	Message string `json:"message" example:"删除集群成功"`
}

// ClusterHealthResponse 集群健康检查响应
type ClusterHealthResponse struct {
	Message string `json:"message" example:"健康检查成功"`
	Data    struct {
		Status  string `json:"status" example:"healthy"`
		Healthy bool   `json:"healthy" example:"true"`
		Error   string `json:"error,omitempty" example:""`
	} `json:"data"`
}

// Add 添加集群
// @Summary 添加集群
// @Description 注册新的Kubernetes集群，支持kubeconfig和Token两种认证方式。如果是kubeconfig，则从YAML中解析Server地址；如果是Token，则需手动指定URL和Token。
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param request body service.CreateRequest true "集群信息"
// @Success 200 {object} ClusterCreateResponse "成功"
// @Failure 400 {object} Response "参数错误或连接失败"
// @Security BearerAuth
// @Router /k8s/cluster/create [post]
func ClusterCreate(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		AuthType   string `json:"authType" binding:"required,oneof=kubeconfig token"`
		Kubeconfig string `json:"kubeconfig"` // 当 authType=kubeconfig 时必填
		Url        string `json:"url"`        // 当 authType=token 时必填
		Token      string `json:"token"`      // 当 authType=token 时必填
		CaData     string `json:"caData"`     // 可选
		Remark     string `json:"remark"`
		Labels     string `json:"labels"`
		Env        string `json:"env" binding:"omitempty,oneof=dev test prod"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 业务逻辑处理
	createReq := &service.CreateRequest{
		Name:       req.Name,
		AuthType:   req.AuthType,
		Kubeconfig: req.Kubeconfig,
		Url:        req.Url,
		Token:      req.Token,
		CaData:     req.CaData,
		Remark:     req.Remark,
		Labels:     req.Labels,
		Env:        req.Env,
	}

	result, err := getService().Create(createReq)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "添加集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加集群成功",
		"data": map[string]interface{}{
			"id":       result.ID,
			"name":     result.Name,
			"url":      result.Url,
			"authType": result.AuthType,
			"status":   result.Status,
		},
	})
}

// Update 更新集群
// @Summary 更新集群
// @Description 更新已注册的Kubernetes集群信息，如果更新了认证信息会自动触发健康检查。仅支持部分字段更新。
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param request body service.UpdateRequest true "集群信息"
// @Success 200 {object} ClusterCreateResponse "成功"
// @Failure 400 {object} Response "参数错误或更新失败"
// @Security BearerAuth
// @Router /k8s/cluster/update [post]
func ClusterUpdate(c *gin.Context) {
	var req struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		Kubeconfig string `json:"kubeconfig"` // 更新 kubeconfig（kubeconfig 模式）
		Url        string `json:"url"`        // 更新 URL（token 模式）
		Token      string `json:"token"`      // 更新 token（token 模式）
		CaData     string `json:"caData"`     // 更新 CA（token 模式）
		Remark     string `json:"remark"`
		Labels     string `json:"labels"`
		Env        string `json:"env" binding:"omitempty,oneof=dev test prod"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 支持从 Path 获取 ID
	if req.ID == 0 {
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				req.ID = uint(id)
			}
		}
	}

	if req.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: id 不能为空",
		})
		return
	}

	updateReq := &service.UpdateRequest{
		ID:         req.ID,
		Name:       req.Name,
		Kubeconfig: req.Kubeconfig,
		Url:        req.Url,
		Token:      req.Token,
		CaData:     req.CaData,
		Remark:     req.Remark,
		Labels:     req.Labels,
		Env:        req.Env,
	}

	result, err := getService().Update(updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "更新集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新集群成功",
		"data": map[string]interface{}{
			"id":     result.ID,
			"name":   result.Name,
			"url":    result.Url,
			"status": result.Status,
		},
	})
}

// Delete 删除集群
// @Summary 删除集群
// @Description 从数据库中删除已注册的Kubernetes集群信息
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param request body object true "集群ID" example({"id": 1})
// @Success 200 {object} ClusterDeleteResponse "成功"
// @Failure 400 {object} Response "参数错误或删除失败"
// @Security BearerAuth
// @Router /k8s/cluster/delete [post]
func ClusterDelete(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := getService().Delete(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "删除集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除集群成功",
	})
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查集群连接状态并更新数据库中的状态字段
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param request body object true "集群ID" example({"id": 1})
// @Success 200 {object} ClusterHealthResponse "成功"
// @Failure 400 {object} Response "参数错误"
// @Security BearerAuth
// @Router /k8s/cluster/health [get]
func ClusterHealthCheck(c *gin.Context) {
	var req struct {
		ID uint `form:"id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	status, err := getService().HealthCheck(req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"success": false,
			"message": "健康检查失败: " + err.Error(),
			"data": map[string]interface{}{
				"status":  status,
				"healthy": false,
				"error":   err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"message": "健康检查成功",
		"data": map[string]interface{}{
			"status":  status,
			"healthy": true,
		},
	})
}

// Search 查找集群
// @Summary 查找集群
// @Description 根据名称或环境快速搜索集群
// @Tags 集群管理
// @Accept json
// @Produce json
// @Param name query string false "集群名称"
// @Param env query string false "环境"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} ClusterListResponse "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Security BearerAuth
// @Router /k8s/cluster/search [get]
func ClusterSearch(c *gin.Context) {
	name := c.Query("name")
	env := c.Query("env")

	if name == "" && env == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "name 或 env 至少传一个",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	clusters, total, err := getService().Search(name, env, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询集群失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "查询成功",
		"data":     clusters,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
