package api

import (
	"net/http"
	"time"

	"devops-platform/internal/modules/monitor/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
)

var monitorService = service.NewMonitorService()

// QueryMonitors godoc
// @Summary 查询监控指标
// @Description 按时间范围查询监控指标数据
// @Tags 监控管理
// @Produce json
// @Security BearerAuth
// @Param metric query string true "指标名"
// @Param start query string false "开始时间(RFC3339)"
// @Param end query string false "结束时间(RFC3339)"
// @Param step query string false "步长"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /monitor/query [get]
func QueryMonitors(c *gin.Context) {
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	result, err := monitorService.Query(
		c.Query("metric"),
		start,
		end,
		c.DefaultQuery("step", "1m"),
	)
	if err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetPrometheusConfig godoc
// @Summary 获取Prometheus配置
// @Description 获取监控模块Prometheus配置
// @Tags 监控管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /monitor/config [get]
func GetPrometheusConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    monitorService.GetConfig(),
	})
}

// SavePrometheusConfig godoc
// @Summary 保存Prometheus配置
// @Description 创建或更新Prometheus配置
// @Tags 监控管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Prometheus配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /monitor/config/upsert [post]
func SavePrometheusConfig(c *gin.Context) {
	var req service.SavePrometheusConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("PROMETHEUS_INVALID_REQUEST", "monitor.SavePrometheusConfig", "参数错误", err))
		return
	}
	data, err := monitorService.SaveConfig(req)
	if err != nil {
		writeObservableError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
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
