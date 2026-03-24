package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/cicd/service"
	"devops-platform/internal/pkg/obserr"

	"github.com/gin-gonic/gin"
)

var cicdService = service.NewCICDService()

// ListPipelines godoc
// @Summary 获取流水线列表
// @Description 获取流水线列表
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/list [get]
func ListPipelines(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.ListPipelineStatus(c.Query("status"), c.Query("keyword")),
	})
}

// ListPipelineStatus godoc
// @Summary 获取流水线状态
// @Description 获取流水线状态统计与列表
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/status [get]
func ListPipelineStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.ListPipelineStatus(c.Query("status"), c.Query("keyword")),
	})
}

// ListPipelineLogs godoc
// @Summary 获取流水线日志
// @Description 获取指定流水线构建日志
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param pipelineId query int false "流水线ID"
// @Param stage query string false "阶段"
// @Param limit query int false "返回数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/logs [get]
func ListPipelineLogs(c *gin.Context) {
	pipelineID, _ := strconv.ParseUint(c.Query("pipelineId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.ListPipelineLogs(uint(pipelineID), c.Query("stage"), limit),
	})
}

// ListPipelineTemplates godoc
// @Summary 获取流水线模板
// @Description 获取流水线模板列表
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/templates [get]
func ListPipelineTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.ListTemplates(c.Query("keyword")),
	})
}

// SavePipelineTemplate godoc
// @Summary 保存流水线模板
// @Description 创建或更新流水线模板
// @Tags CI/CD管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "模板信息"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /cicd/template/save [post]
func SavePipelineTemplate(c *gin.Context) {
	var req service.SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("JENKINS_INVALID_REQUEST", "cicd.SavePipelineTemplate", "参数错误", err))
		return
	}
	data, err := cicdService.SaveTemplate(req)
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

// PreviewPipelineOrchestration godoc
// @Summary 预览流水线编排
// @Description 预览流水线在目标环境的编排结果
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param pipelineId query int false "流水线ID"
// @Param templateId query int false "模板ID"
// @Param environment query string false "环境"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /cicd/orchestration/preview [get]
func PreviewPipelineOrchestration(c *gin.Context) {
	pipelineID, _ := strconv.ParseUint(c.Query("pipelineId"), 10, 64)
	templateID, _ := strconv.ParseUint(c.Query("templateId"), 10, 64)
	data, err := cicdService.PreviewOrchestration(uint(pipelineID), uint(templateID), c.Query("environment"), nil)
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

// TriggerPipeline godoc
// @Summary 触发流水线
// @Description 手动触发流水线执行
// @Tags CI/CD管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "触发参数"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /cicd/trigger [post]
func TriggerPipeline(c *gin.Context) {
	var req service.TriggerPipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("JENKINS_INVALID_REQUEST", "cicd.TriggerPipeline", "参数错误", err))
		return
	}
	data, err := cicdService.TriggerPipeline(req)
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

// ListPipelineRuns godoc
// @Summary 获取流水线运行记录
// @Description 获取流水线运行历史
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Param pipelineId query int false "流水线ID"
// @Param status query string false "状态"
// @Param limit query int false "返回数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/runs [get]
func ListPipelineRuns(c *gin.Context) {
	pipelineID, _ := strconv.ParseUint(c.Query("pipelineId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.ListPipelineRuns(uint(pipelineID), c.Query("status"), limit),
	})
}

// GetJenkinsConfig godoc
// @Summary 获取Jenkins配置
// @Description 获取CI/CD模块Jenkins配置
// @Tags CI/CD管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /cicd/config [get]
func GetJenkinsConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cicdService.GetConfig(),
	})
}

// SaveJenkinsConfig godoc
// @Summary 保存Jenkins配置
// @Description 创建或更新Jenkins配置
// @Tags CI/CD管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Jenkins配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /cicd/config/upsert [post]
func SaveJenkinsConfig(c *gin.Context) {
	var req service.SaveJenkinsConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeObservableError(c, http.StatusBadRequest, obserr.Wrap("JENKINS_INVALID_REQUEST", "cicd.SaveJenkinsConfig", "参数错误", err))
		return
	}
	data, err := cicdService.SaveConfig(req)
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
