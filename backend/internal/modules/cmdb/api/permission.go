package api

import (
	"errors"
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PermissionList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("userId", "0"), 10, 32)
	hostGroupID, _ := strconv.ParseUint(c.DefaultQuery("hostGroupId", "0"), 10, 32)
	permission := c.Query("permission")

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	perms, total, err := svc.ListInTenant(tenantID, page, pageSize, uint(userID), uint(hostGroupID), permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取权限列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     perms,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func PermissionCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	currentUserID := userIDValue.(uint)

	var req service.PermissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	created, err := svc.CreateInTenant(tenantID, currentUserID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "授予权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "授权成功",
		"data":    created,
	})
}

func PermissionUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.PermissionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	perm, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    perm,
	})
}

func PermissionDelete(c *gin.Context) {
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

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

func PermissionMyHosts(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID := userIDValue.(uint)

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	hosts, err := svc.MyHosts(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    hosts,
	})
}

func PermissionCheck(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	userID := userIDValue.(uint)

	hostID, _ := strconv.ParseUint(c.DefaultQuery("hostId", "0"), 10, 32)
	action := c.DefaultQuery("action", "view")
	if hostID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 hostId 参数"})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	allowed, perm, err := svc.CheckPermission(tenantID, userID, uint(hostID), action)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "检查权限失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"allowed":    allowed,
			"permission": perm,
		},
	})
}

func PermissionGroupHostCount(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	groupID, _ := strconv.ParseUint(c.DefaultQuery("groupId", "0"), 10, 32)
	if groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少 groupId 参数"})
		return
	}

	svc := getPermissionService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	count, err := svc.GetGroupHostCount(tenantID, uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机数量失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"hostCount": count,
		},
	})
}
