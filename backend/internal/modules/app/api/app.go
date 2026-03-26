package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/app/model"
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

// ========== 应用配置相关 API ==========

// GetAppConfig godoc
// @Summary 获取应用配置
// @Description 获取应用的完整配置信息
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/config [get]
func GetAppConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	config, err := appService.GetAppConfig(uint(appID))
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
		"data":    config,
	})
}

// SaveAppConfig godoc
// @Summary 保存应用配置
// @Description 保存应用的基础配置信息
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.AppConfig true "应用配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/config [post]
func SaveAppConfig(c *gin.Context) {
	var config model.AppConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.ValidateAppConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveAppConfig(config)
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

// ========== 构建配置相关 API ==========

// GetBuildConfig godoc
// @Summary 获取构建配置
// @Description 获取应用的构建配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/build-config [get]
func GetBuildConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	config, err := appService.GetBuildConfig(uint(appID))
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
		"data":    config,
	})
}

// SaveBuildConfig godoc
// @Summary 保存构建配置
// @Description 保存应用的构建配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.BuildConfig true "构建配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-config [post]
func SaveBuildConfig(c *gin.Context) {
	var config model.BuildConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.ValidateBuildConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveBuildConfig(config)
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

// ========== 部署配置相关 API ==========

// GetDeployConfig godoc
// @Summary 获取部署配置
// @Description 获取应用的部署配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/deploy-config [get]
func GetDeployConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	config, err := appService.GetDeployConfig(uint(appID))
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
		"data":    config,
	})
}

// SaveDeployConfig godoc
// @Summary 保存部署配置
// @Description 保存应用的部署配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeployConfig true "部署配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-config [post]
func SaveDeployConfig(c *gin.Context) {
	var config model.DeployConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.ValidateDeployConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveDeployConfig(config)
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

// ========== 技术栈配置相关 API ==========

// GetTechStackConfig godoc
// @Summary 获取技术栈配置
// @Description 获取应用的技术栈配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/tech-stack [get]
func GetTechStackConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	config, err := appService.GetTechStackConfig(uint(appID))
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
		"data":    config,
	})
}

// SaveTechStackConfig godoc
// @Summary 保存技术栈配置
// @Description 保存应用的技术栈配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.TechStackConfig true "技术栈配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/tech-stack [post]
func SaveTechStackConfig(c *gin.Context) {
	var config model.TechStackConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.ValidateTechStackConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveTechStackConfig(config)
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

// ========== 枚举值相关 API ==========

// GetEnumOptions godoc
// @Summary 获取枚举选项
// @Description 获取预定义的枚举选项（应用状态、构建环境、语言、环境等）
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/enums [get]
func GetEnumOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"appStates": []gin.H{
				{"label": "运行中", "value": model.AppStateRunning},
				{"label": "已停止", "value": model.AppStateStopped},
				{"label": "开发中", "value": model.AppStateDeveloping},
			},
			"buildEnvs": []gin.H{
				{"label": "开发环境", "value": model.BuildEnvDevelopment},
				{"label": "预发布环境", "value": model.BuildEnvStaging},
				{"label": "生产环境", "value": model.BuildEnvProduction},
			},
			"languages": []gin.H{
				{"label": "Java", "value": model.LanguageJava},
				{"label": "Go", "value": model.LanguageGo},
				{"label": "Python", "value": model.LanguagePython},
				{"label": "NodeJS", "value": model.LanguageNodeJS},
			},
			"environments": []gin.H{
				{"label": "开发", "value": model.EnvironmentDev},
				{"label": "测试", "value": model.EnvironmentTest},
				{"label": "预发布", "value": model.EnvironmentStaging},
				{"label": "生产", "value": model.EnvironmentProd},
			},
			"cpuOptions":    model.GetCPUOptions(),
			"memoryOptions": model.GetMemoryOptions(),
			"defaultTechStacks": gin.H{
				"java":   service.GetDefaultTechStack(model.LanguageJava),
				"go":     service.GetDefaultTechStack(model.LanguageGo),
				"python": service.GetDefaultTechStack(model.LanguagePython),
			},
			"defaultBuildConfigs": gin.H{
				"java":   service.GetDefaultBuildConfig(model.LanguageJava),
				"go":     service.GetDefaultBuildConfig(model.LanguageGo),
				"python": service.GetDefaultBuildConfig(model.LanguagePython),
			},
			"defaultDeployConfig": service.GetDefaultDeployConfig(),
		},
	})
}

// ========== 删除配置相关 API ==========

// DeleteAppConfig godoc
// @Summary 删除应用配置
// @Description 删除应用的基础配置信息
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/config [delete]
func DeleteAppConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	if !appService.DeleteAppConfig(uint(appID)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "应用配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// DeleteBuildConfig godoc
// @Summary 删除构建配置
// @Description 删除应用的构建配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/build-config [delete]
func DeleteBuildConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	if !appService.DeleteBuildConfig(uint(appID)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "构建配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// DeleteDeployConfig godoc
// @Summary 删除部署配置
// @Description 删除应用的部署配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/deploy-config [delete]
func DeleteDeployConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	if !appService.DeleteDeployConfig(uint(appID)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "部署配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// DeleteTechStackConfig godoc
// @Summary 删除技术栈配置
// @Description 删除应用的技术栈配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param appId path int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/{appId}/tech-stack [delete]
func DeleteTechStackConfig(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("appId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "appId参数错误",
		})
		return
	}

	if !appService.DeleteTechStackConfig(uint(appID)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "技术栈配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}
