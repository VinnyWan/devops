package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"devops-platform/internal/modules/cmdb/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HostList 主机列表
func HostList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	groupID, _ := strconv.ParseUint(c.DefaultQuery("groupId", "0"), 10, 64)
	status := c.Query("status")
	keyword := c.Query("keyword")

	// 主机级权限过滤
	var allowedHostIDs []uint
	userID := c.GetUint("userID")
	if !isCmdbAdmin(c, tenantID, userID) {
		permSvc := getPermissionService()
		hosts, err := permSvc.MyHosts(tenantID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取主机权限失败"})
			return
		}
		allowedHostIDs = make([]uint, 0, len(hosts))
		for _, h := range hosts {
			allowedHostIDs = append(allowedHostIDs, h.HostID)
		}
		if len(allowedHostIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"list": []interface{}{}, "total": 0}})
			return
		}
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	hosts, total, err := svc.ListInTenant(tenantID, page, pageSize, uint(groupID), status, keyword, allowedHostIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     hosts,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// HostStats 主机统计
func HostStats(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	stats, err := svc.StatsInTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取统计失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// HostDetail 主机详情
func HostDetail(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的主机 ID"})
		return
	}

	// 主机级权限校验
	userID := c.GetUint("userID")
	if !isCmdbAdmin(c, tenantID, userID) {
		permSvc := getPermissionService()
		allowed, _, err := permSvc.CheckPermission(tenantID, userID, uint(id), "view")
		if err != nil || !allowed {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "主机不存在或无访问权限"})
			return
		}
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	host, err := svc.GetByIDInTenant(tenantID, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "主机不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取主机详情失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    host,
	})
}

// HostCreate 创建主机
func HostCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.HostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	host, err := svc.CreateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "创建主机失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    host,
	})
}

// HostBatchCreate 批量导入主机
func HostBatchCreate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var reqs []service.HostCreateRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "导入列表不能为空"})
		return
	}
	if len(reqs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "单次导入不能超过 100 条"})
		return
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	hosts, err := svc.BatchCreateInTenant(tenantID, reqs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "批量导入失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "导入成功",
		"data":    hosts,
		"total":   len(hosts),
	})
}

// HostUpdate 更新主机
func HostUpdate(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req service.HostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	host, err := svc.UpdateInTenant(tenantID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新主机失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    host,
	})
}

// HostDelete 删除主机
func HostDelete(c *gin.Context) {
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

	svc := getHostService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	if err := svc.DeleteInTenant(tenantID, req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "删除主机失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// HostTest 测试主机连接
func HostTest(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	var req struct {
		ID           uint `json:"id"`
		CredentialID uint `json:"credentialId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "error": err.Error()})
		return
	}

	svc := getHostService()
	credSvc := getCredentialService()
	if svc == nil || credSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	host, err := svc.GetByIDInTenant(tenantID, req.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "主机不存在"})
		return
	}

	credID := req.CredentialID
	if credID == 0 && host.CredentialID != nil {
		credID = *host.CredentialID
	}
	if credID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "未指定凭据"})
		return
	}

	cred, err := credSvc.GetDecryptedInTenant(tenantID, credID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "凭据不存在"})
		return
	}

	ok, msg := svc.TestConnection(host.Ip, host.Port, cred)
	status := "online"
	if !ok {
		status = "offline"
	}

	updates := map[string]interface{}{
		"status": status,
	}
	if ok {
		now := time.Now()
		updates["last_active_at"] = &now
	}
	if cmdbDB != nil {
		_ = cmdbDB.Model(host).Updates(updates).Error
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": msg,
		"data": gin.H{
			"connected": ok,
			"status":    status,
			"message":   msg,
		},
	})
}
