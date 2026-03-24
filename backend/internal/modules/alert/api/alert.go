package api

import (
	"net/http"
	"strconv"
	"time"

	"devops-platform/internal/modules/alert/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
)

var alertService = service.NewAlertService()

// ListAlertRules godoc
// @Summary 获取告警规则列表
// @Description 按关键词筛选告警规则
// @Tags 告警管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /alert/rules [get]
func ListAlertRules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alertService.ListRules(c.Query("keyword")),
	})
}

// ListAlertHistory godoc
// @Summary 获取告警历史
// @Description 按状态和时间范围获取告警历史
// @Tags 告警管理
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态"
// @Param start query string false "开始时间(RFC3339)"
// @Param end query string false "结束时间(RFC3339)"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /alert/history [get]
func ListAlertHistory(c *gin.Context) {
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alertService.ListHistory(c.Query("status"), start, end),
	})
}

// ToggleAlertRule godoc
// @Summary 启停告警规则
// @Description 启用或禁用告警规则
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "规则启停参数"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "规则不存在"
// @Router /alert/rule/toggle [post]
func ToggleAlertRule(c *gin.Context) {
	var req service.RuleEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("ALERT_INVALID_REQUEST", "alert.ToggleAlertRule", "参数错误", err))
		return
	}
	if req.ID == 0 {
		writeObservableError(c, http.StatusBadRequest, obserr.New("ALERT_RULE_ID_REQUIRED", "alert.ToggleAlertRule", "rule id 不能为空"))
		return
	}
	rule, err := alertService.SetRuleEnabled(req)
	if err != nil {
		status := http.StatusBadRequest
		if alertService.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeObservableError(c, status, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    rule,
	})
}

// ListAlertSilences godoc
// @Summary 获取告警静默列表
// @Description 查询告警规则静默配置
// @Tags 告警管理
// @Produce json
// @Security BearerAuth
// @Param ruleId query int false "规则ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /alert/silences [get]
func ListAlertSilences(c *gin.Context) {
	ruleID, _ := strconv.ParseUint(c.Query("ruleId"), 10, 64)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alertService.ListSilences(uint(ruleID)),
	})
}

// UpsertAlertSilence godoc
// @Summary 保存告警静默
// @Description 创建或更新告警静默配置
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "静默配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /alert/silence/upsert [post]
func UpsertAlertSilence(c *gin.Context) {
	var req service.SilenceUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("ALERT_INVALID_REQUEST", "alert.UpsertAlertSilence", "参数错误", err))
		return
	}
	data, err := alertService.UpsertSilence(req)
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

// ListAlertChannels godoc
// @Summary 获取通知渠道列表
// @Description 按类型筛选告警通知渠道
// @Tags 告警管理
// @Produce json
// @Security BearerAuth
// @Param type query string false "渠道类型"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /alert/channels [get]
func ListAlertChannels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alertService.ListChannels(c.Query("type")),
	})
}

// UpsertAlertChannel godoc
// @Summary 保存通知渠道
// @Description 创建或更新告警通知渠道
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "渠道配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /alert/channel/upsert [post]
func UpsertAlertChannel(c *gin.Context) {
	var req service.ChannelUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("ALERT_INVALID_REQUEST", "alert.UpsertAlertChannel", "参数错误", err))
		return
	}
	data, err := alertService.UpsertChannel(req)
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

// GetAlertmanagerConfig godoc
// @Summary 获取Alertmanager配置
// @Description 获取告警模块Alertmanager配置
// @Tags 告警管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /alert/config [get]
func GetAlertmanagerConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alertService.GetConfig(),
	})
}

// SaveAlertmanagerConfig godoc
// @Summary 保存Alertmanager配置
// @Description 创建或更新Alertmanager配置
// @Tags 告警管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Alertmanager配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /alert/config/upsert [post]
func SaveAlertmanagerConfig(c *gin.Context) {
	var req service.SaveAlertmanagerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("ALERT_INVALID_REQUEST", "alert.SaveAlertmanagerConfig", "参数错误", err))
		return
	}
	data, err := alertService.SaveConfig(req)
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
	msg, _ := details["message"].(string)
	code, _ := details["code"].(string)
	c.JSON(status, gin.H{
		"code":    status,
		"message": msg,
		"error": gin.H{
			"code":  code,
			"chain": details["chain"],
		},
	})
}
