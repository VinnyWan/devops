package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	auditService *service.AuditService
	auditOnce    sync.Once
)

func getAuditService() *service.AuditService {
	auditOnce.Do(func() {
		auditService = service.NewAuditService(repository.NewAuditRepo(db))
	})
	return auditService
}

// ListAuditLogs godoc
// @Summary 获取审计日志列表
// @Description 按条件分页查询审计日志
// @Tags 审计管理
// @Produce json
// @Security BearerAuth
// @Param userId query int false "用户ID"
// @Param username query string false "用户名"
// @Param operation query string false "操作类型"
// @Param resource query string false "资源路径"
// @Param keyword query string false "关键词（长度>=3生效，跨字段匹配）"
// @Param startAt query string false "开始时间"
// @Param endAt query string false "结束时间"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "查询失败"
// @Router /audit/list [get]
func ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	userID, hasUserID, err := parseUintQuery(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userId 参数非法",
		})
		return
	}

	req := service.AuditListRequest{
		Username:  c.Query("username"),
		Operation: c.Query("operation"),
		Resource:  c.Query("resource"),
		Keyword:   c.Query("keyword"),
		StartAt:   c.Query("startAt"),
		EndAt:     c.Query("endAt"),
		Page:      page,
		PageSize:  pageSize,
	}
	if hasUserID {
		req.UserID = &userID
	}

	logs, total, err := getAuditService().List(req)
	if err != nil {
		logger.Log.Error("Failed to list audit logs", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "查询审计日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// ExportAuditLogs godoc
// @Summary 导出审计日志CSV
// @Description 按条件导出审计日志CSV
// @Tags 审计管理
// @Produce text/csv
// @Security BearerAuth
// @Param userId query int false "用户ID"
// @Param username query string false "用户名"
// @Param operation query string false "操作类型"
// @Param resource query string false "资源路径"
// @Param keyword query string false "关键词（长度>=3生效，跨字段匹配）"
// @Param startAt query string false "开始时间"
// @Param endAt query string false "结束时间"
// @Param limit query int false "导出上限"
// @Success 200 {string} string "CSV内容"
// @Failure 400 {object} map[string]interface{} "导出失败"
// @Router /audit/export [get]
func ExportAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10000"))
	userID, hasUserID, err := parseUintQuery(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "userId 参数非法",
		})
		return
	}
	req := service.AuditListRequest{
		Username:  c.Query("username"),
		Operation: c.Query("operation"),
		Resource:  c.Query("resource"),
		Keyword:   c.Query("keyword"),
		StartAt:   c.Query("startAt"),
		EndAt:     c.Query("endAt"),
	}
	if hasUserID {
		req.UserID = &userID
	}

	logs, err := getAuditService().Export(req, limit)
	if err != nil {
		logger.Log.Error("Failed to export audit logs", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "导出审计日志失败: " + err.Error(),
		})
		return
	}
	filename := "audit_logs_" + time.Now().Format("20060102150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	writer := csv.NewWriter(c.Writer)
	if err := writer.Write([]string{"ID", "用户ID", "用户名", "操作类型", "HTTP方法", "资源路径", "状态码", "耗时(ms)", "IP", "请求时间", "创建时间"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "写入CSV失败"})
		return
	}
	for _, item := range logs {
		record := []string{
			toString(item["id"]),
			toString(item["userId"]),
			toString(item["username"]),
			toString(item["operation"]),
			toString(item["method"]),
			toString(item["path"]),
			toString(item["status"]),
			toString(item["latency"]),
			toString(item["ip"]),
			toString(item["requestAt"]),
			toString(item["createdAt"]),
		}
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "写入CSV失败"})
			return
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "导出CSV失败"})
	}
}

// CleanupExpiredAuditLogs godoc
// @Summary 手动清理过期审计日志
// @Description 立即执行一次过期审计日志清理
// @Tags 审计管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 500 {object} map[string]interface{} "清理失败"
// @Router /audit/cleanup [post]
func CleanupExpiredAuditLogs(c *gin.Context) {
	affected, err := getAuditService().CleanExpiredNow()
	if err != nil {
		logger.Log.Error("Failed to cleanup audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理过期审计日志失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "清理成功",
		"data": gin.H{
			"cleaned": affected,
		},
	})
}

func parseUintQuery(raw string) (uint, bool, error) {
	if raw == "" {
		return 0, false, nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return uint(id), true, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return typed
	case time.Time:
		return typed.Format(time.RFC3339)
	default:
		return fmt.Sprint(v)
	}
}
