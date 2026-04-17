package api

import (
	"errors"
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GroupTree 分组树
func GroupTreeAPI(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	svc := getGroupService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	tree, err := svc.GetTreeInTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取分组树失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    tree,
	})
}

// GroupDetail 分组详情
func GroupDetail(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的 ID"})
		return
	}

	svc := getGroupService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	group, err := svc.GetByIDInTenant(tenantID, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "分组不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取分组详情失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    group,
	})
}

// GroupCreate 创建分组
func GroupCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.GroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getGroupService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	group, err := svc.CreateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "创建分组失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    group,
	})
}

// GroupUpdate 更新分组
func GroupUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.GroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getGroupService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	group, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新分组失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    group,
	})
}

// GroupDelete 删除分组
func GroupDelete(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getGroupService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除分组失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
