package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devops-platform/internal/modules/log/model"
	"devops-platform/internal/modules/log/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var logSvc *service.LogService

func SetLogDB(db *gorm.DB) {
	logSvc = service.NewLogService(db)
}

// ListLogSources godoc
// @Summary 获取日志源列表
// @Description 分页获取日志源列表
// @Tags 日志管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/sources [get]
func ListLogSources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	sources, total, err := logSvc.ListSources(page, pageSize)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": sources, "total": total})
}

// SaveLogSource godoc
// @Summary 保存日志源
// @Description 创建或更新日志源
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.LogSource true "日志源信息"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/sources [post]
func SaveLogSource(c *gin.Context) {
	var src model.LogSource
	if err := c.ShouldBindJSON(&src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	if err := logSvc.SaveSource(&src); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": src})
}

// DeleteLogSource godoc
// @Summary 删除日志源
// @Description 根据ID删除日志源
// @Tags 日志管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志源ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/sources/{id} [delete]
func DeleteLogSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := logSvc.DeleteSource(uint(id)); err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

// TestLogSourceConnection godoc
// @Summary 测试日志源连接
// @Description 测试指定日志源是否连通
// @Tags 日志管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志源ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/sources/{id}/test [post]
func TestLogSourceConnection(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := logSvc.TestConnection(uint(id)); err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "connection successful"})
}

// SearchLogs godoc
// @Summary 检索日志
// @Description 按条件分页检索日志
// @Tags 日志管理
// @Produce json
// @Security BearerAuth
// @Param request body model.SearchRequest true "检索条件"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/search [post]
func SearchLogs(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	resp, err := logSvc.Search(req)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}

// ExportLogs godoc
// @Summary 导出日志
// @Description 按条件导出日志为CSV
// @Tags 日志管理
// @Produce text/csv
// @Security BearerAuth
// @Param request body model.SearchRequest true "导出条件"
// @Success 200 {file} file "CSV文件"
// @Router /log/export [post]
func ExportLogs(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}
	req.Page = 1
	req.PageSize = 10000
	resp, err := logSvc.Search(req)
	if err != nil {
		writeObservableError(c, http.StatusInternalServerError, err)
		return
	}

	// Write CSV
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=logs.csv")
	c.Writer.WriteString("Timestamp,Level,Service,Host,Message\n")
	for _, e := range resp.Entries {
		c.Writer.WriteString(fmt.Sprintf("%s,%s,%s,%s,\"%s\"\n",
			e.Timestamp, e.Level, e.Service, e.Host, strings.ReplaceAll(e.Message, "\"", "\"\"")))
	}
}

func writeObservableError(c *gin.Context, status int, err error) {
	details := obserr.Details(err)
	code, _ := details["code"].(string)
	msg, _ := details["message"].(string)
	c.JSON(status, gin.H{
		"code":    status,
		"message": msg,
		"error": gin.H{
			"code":  code,
			"chain": details["chain"],
		},
	})
}
