package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/app/service"

	"github.com/gin-gonic/gin"
)

var appService = service.NewAppService()

// ListApps godoc
// @Summary 获取应用列表
// @Description 获取应用列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/list [get]
func ListApps(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.List(),
	})
}

// CreateApp godoc
// @Summary 保存应用模板
// @Description 创建或更新应用模板
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "模板信息"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/template/save [post]
func CreateApp(c *gin.Context) {
	var req service.SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.SaveTemplate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// UpdateApp godoc
// @Summary 部署应用
// @Description 触发应用部署
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "部署参数"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy [post]
func UpdateApp(c *gin.Context) {
	var req service.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.Deploy(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// DeleteApp godoc
// @Summary 回滚应用
// @Description 执行应用版本回滚
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "回滚参数"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/rollback [post]
func DeleteApp(c *gin.Context) {
	var req service.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.Rollback(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// ListTemplates godoc
// @Summary 获取应用模板列表
// @Description 获取应用模板列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/template/list [get]
func ListTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListTemplates(c.Query("keyword")),
	})
}

// ListDeployments godoc
// @Summary 获取部署记录
// @Description 获取应用部署记录列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId query int false "应用ID"
// @Param environment query string false "环境"
// @Param limit query int false "返回数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/deployment/list [get]
func ListDeployments(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListDeployments(uint(appID), c.Query("environment"), limit),
	})
}

// ListVersions godoc
// @Summary 获取版本列表
// @Description 获取应用版本列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId query int false "应用ID"
// @Param limit query int false "返回数量"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/version/list [get]
func ListVersions(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListVersions(uint(appID), limit),
	})
}

// QueryTopology godoc
// @Summary 查询应用拓扑
// @Description 按应用和环境查询拓扑信息
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId query int false "应用ID"
// @Param environment query string false "环境"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/topology [get]
func QueryTopology(c *gin.Context) {
	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	data, err := appService.QueryTopology(uint(appID), c.Query("environment"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}
