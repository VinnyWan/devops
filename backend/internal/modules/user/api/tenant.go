package api

import (
	"net/http"
	"strconv"
	"sync"

	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	tenantService *service.TenantService
	tenantOnce    sync.Once
)

func getTenantService() *service.TenantService {
	tenantOnce.Do(func() {
		tenantService = service.NewTenantService(repository.NewTenantRepo(db))
	})
	return tenantService
}

// ListTenants godoc
// @Summary 租户列表
// @Description 获取租户列表
// @Tags 租户管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "关键词"
// @Param status query string false "状态"
// @Router /tenant/list [get]
func ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	items, total, err := getTenantService().List(page, pageSize, keyword, status)
	if err != nil {
		logger.Log.Error("Failed to list tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取租户列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  items,
			"total": total,
			"page":  page,
		},
	})
}

// GetTenantDetail godoc
// @Summary 租户详情
// @Description 根据ID查询租户详情
// @Tags 租户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "租户ID"
// @Router /tenant/detail [get]
func GetTenantDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的租户ID"})
		return
	}

	item, err := getTenantService().GetByID(uint(id))
	if err != nil {
		logger.Log.Error("Failed to get tenant detail", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "租户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": item})
}

// CreateTenant godoc
// @Summary 创建租户
// @Description 创建新租户
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateTenantRequest true "租户信息"
// @Router /tenant/create [post]
func CreateTenant(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	item, err := getTenantService().Create(&req)
	if err != nil {
		logger.Log.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": item})
}

// UpdateTenant godoc
// @Summary 更新租户
// @Description 更新租户信息
// @Tags 租户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateTenantRequest true "租户信息"
// @Router /tenant/update [post]
func UpdateTenant(c *gin.Context) {
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	if err := getTenantService().Update(&req); err != nil {
		logger.Log.Error("Failed to update tenant", zap.Uint("tenantID", req.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DisableTenant godoc
// @Summary 停用租户
// @Description 将租户状态置为 inactive
// @Tags 租户管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "租户ID"
// @Router /tenant/disable [post]
func DisableTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的租户ID"})
		return
	}

	if err := getTenantService().Disable(uint(id)); err != nil {
		logger.Log.Error("Failed to disable tenant", zap.Uint64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "停用成功"})
}
