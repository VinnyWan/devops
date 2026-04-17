package api

import (
	"errors"
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CredentialList 凭据列表
func CredentialList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	svc := getCredentialService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	creds, total, err := svc.ListInTenant(tenantID, page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取凭据列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     creds,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CredentialDetail 凭据详情（不返回敏感字段）
func CredentialDetail(c *gin.Context) {
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

	svc := getCredentialService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	cred, err := svc.GetByIDInTenant(tenantID, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "凭据不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取凭据详情失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    cred,
	})
}

// CredentialCreate 创建凭据
func CredentialCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.CredentialCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getCredentialService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	cred, err := svc.CreateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "创建凭据失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    cred,
	})
}

// CredentialUpdate 更新凭据
func CredentialUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.CredentialUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getCredentialService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	cred, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新凭据失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    cred,
	})
}

// CredentialDelete 删除凭据
func CredentialDelete(c *gin.Context) {
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

	svc := getCredentialService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除凭据失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
