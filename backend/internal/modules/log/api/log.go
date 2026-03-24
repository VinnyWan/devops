package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/log/model"
	"devops-platform/internal/modules/log/service"

	"github.com/gin-gonic/gin"
)

var logService = service.NewLogService()

// SearchLogs godoc
// @Summary 检索日志
// @Description 按条件分页检索日志
// @Tags 日志管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Param source query string false "来源"
// @Param level query string false "级别"
// @Param start query string false "开始时间(RFC3339)"
// @Param end query string false "结束时间(RFC3339)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /log/search [get]
func SearchLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	result := logService.Search(model.SearchRequest{
		Keyword:  c.Query("keyword"),
		Source:   c.Query("source"),
		Level:    c.Query("level"),
		Start:    service.ParseTime(c.Query("start")),
		End:      service.ParseTime(c.Query("end")),
		Page:     page,
		PageSize: pageSize,
	})
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}
