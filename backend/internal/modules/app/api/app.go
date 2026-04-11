package api

import (
	"net/http"
	"strconv"

	"devops-platform/internal/modules/app/model"
	apprepo "devops-platform/internal/modules/app/repository"
	"devops-platform/internal/modules/app/service"
	k8sapi "devops-platform/internal/modules/k8s/api"

	"github.com/gin-gonic/gin"
)

var sharedAppRepo = apprepo.NewAppRepo()
var appService = service.NewAppServiceWithRepo(sharedAppRepo)
var appMenuService = service.NewAppMenuService()
var buildEnvService = service.NewBuildEnvService()
var containerConfigService = service.NewContainerConfigService()
var enumService = service.NewEnumServiceWithRepo(sharedAppRepo)

func getCurrentTenantID(c *gin.Context) uint {
	if tenantID, exists := c.Get("tenantID"); exists {
		if id, ok := tenantID.(uint); ok {
			return id
		}
	}
	return 0
}

func requireTenantID(c *gin.Context) (uint, bool) {
	tenantID := getCurrentTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未认证：租户上下文缺失",
		})
		return 0, false
	}
	return tenantID, true
}

// ========== 菜单选项 ==========

// GetAppManagementMenuOptions godoc
// @Summary 获取应用管理菜单选项
// @Description 获取应用管理下拉框的4个选项
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/menu/options [get]
func GetAppManagementMenuOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appMenuService.GetAppManagementMenuOptions(),
	})
}

// ========== 1. 应用列表 + 搜索 + 筛选接口（GET）==========

// ListApps godoc
// @Summary 获取应用列表
// @Description 获取应用列表，支持分页、模糊搜索、条件筛选
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1"
// @Param size query int false "每页条数，默认10"
// @Param keyword query string false "搜索关键字（应用名称、负责人模糊匹配）"
// @Param instance_type query string false "实例类型筛选（container/native）"
// @Param status query string false "状态筛选（running/offline）"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/list [get]
func ListApps(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	instanceType := c.Query("instance_type")
	status := c.Query("status")

	list, total := appService.ListAppsWithFilterInTenant(tenantID, page, size, keyword, instanceType, status)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"total": total,
			"list":  list,
		},
	})
}

// ========== 2. 单应用配置详情查询接口（GET）==========

// GetAppConfig godoc
// @Summary 获取应用配置详情
// @Description 根据应用ID查询应用的完整配置信息
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/config [get]
func GetAppConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "id参数错误",
		})
		return
	}

	config, err := appService.GetAppConfigInTenant(tenantID, uint(appID))
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

// ========== 3. 应用配置新增/修改接口（POST）==========

// SaveAppConfig godoc
// @Summary 保存应用配置
// @Description 新增或修改应用配置，id为空时新增，否则修改
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.AppConfig true "应用配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/config/save [post]
func SaveAppConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

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

	data, err := appService.SaveAppConfigInTenant(tenantID, config)
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

// ========== 4. 应用状态切换接口（POST）==========

// ToggleAppStatus godoc
// @Summary 应用状态切换
// @Description 将应用状态在「运行中 ↔ 已下线」之间切换
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.StatusToggleRequest true "状态切换请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/status/toggle [post]
func ToggleAppStatus(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.StatusToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	data, err := appService.ToggleAppStatusInTenant(tenantID, req.ID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "状态切换成功",
		"data":    data,
	})
}

// ========== 5. 构建配置查询接口（GET）==========

// GetBuildConfig godoc
// @Summary 获取构建配置
// @Description 根据应用ID获取构建配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-config [get]
func GetBuildConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	config, err := appService.GetBuildConfigInTenant(tenantID, uint(appID))
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

// ========== 6. 构建配置保存/修改接口（POST）==========

// SaveBuildConfig godoc
// @Summary 保存构建配置
// @Description 新增或更新应用的构建配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.BuildConfig true "构建配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-config/save [post]
func SaveBuildConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

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

	data, err := appService.SaveBuildConfigInTenant(tenantID, config)
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

// ========== 7. 部署配置查询接口（GET）==========

// GetDeployConfig godoc
// @Summary 获取部署配置
// @Description 根据应用ID和环境获取部署配置（k8s集群、副本数）
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Param environment query string false "环境（dev/test/staging/prod），不传则返回第一个配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-config [get]
func GetDeployConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	environment := c.Query("environment")

	var config model.DeployConfig
	if environment != "" {
		config, err = appService.GetDeployConfigInTenant(tenantID, uint(appID), environment)
	} else {
		// 兼容旧接口：不传环境则返回第一个配置
		config, err = appService.GetDeployConfigByAppIDInTenant(tenantID, uint(appID))
	}

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

// ListDeployConfigs godoc
// @Summary 获取应用所有环境的部署配置
// @Description 获取应用所有环境的部署配置列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-configs [get]
func ListDeployConfigs(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	configs := appService.ListDeployConfigsByAppInTenant(tenantID, uint(appID))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    configs,
	})
}

// ========== 8. 部署配置保存/修改接口（POST）==========

// SaveDeployConfig godoc
// @Summary 保存部署配置
// @Description 提交/更新k8s集群、副本数选择（按环境区分）
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeployConfigRequest true "部署配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-config/save [post]
func SaveDeployConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeployConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	config := model.DeployConfig{
		ID:            req.ID,
		AppID:         req.AppID,
		Environment:   req.Environment,
		ClusterName:   req.ClusterName,
		Replicas:      req.Replicas,
		ServicePort:   req.ServicePort,
		CPURequest:    req.CPURequest,
		CPULimit:      req.CPULimit,
		MemoryRequest: req.MemoryRequest,
		MemoryLimit:   req.MemoryLimit,
		EnvVars:       req.EnvVars,
	}

	if err := service.ValidateDeployConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveDeployConfigInTenant(tenantID, config)
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

// ========== 9. 容器配置查询接口（GET）==========

// GetContainerConfig godoc
// @Summary 获取容器配置
// @Description 根据应用ID和环境获取容器配置（CPU/内存、挂载目录、环境变量）
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Param environment query string false "环境（dev/test/staging/prod），不传则返回第一个配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/container-config [get]
func GetContainerConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	environment := c.Query("environment")

	var config model.ContainerConfig
	if environment != "" {
		config, err = appService.GetContainerConfigInTenant(tenantID, uint(appID), environment)
	} else {
		// 兼容旧接口：不传环境则返回第一个配置
		config, err = appService.GetContainerConfigByAppIDInTenant(tenantID, uint(appID))
	}

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

// ListContainerConfigs godoc
// @Summary 获取应用所有环境的容器配置
// @Description 获取应用所有环境的容器配置列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/container-configs [get]
func ListContainerConfigs(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	configs := appService.ListContainerConfigsByAppInTenant(tenantID, uint(appID))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    configs,
	})
}

// ========== 10. 容器配置保存/修改接口（POST）==========

// SaveContainerConfig godoc
// @Summary 保存容器配置
// @Description 提交/更新CPU/内存、镜像、挂载目录、环境变量KV配置（按环境区分）
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ContainerConfigRequest true "容器配置"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/container-config/save [post]
func SaveContainerConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.ContainerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	config := model.ContainerConfig{
		ID:            req.ID,
		AppID:         req.AppID,
		Environment:   req.Environment,
		Namespace:     req.Namespace,
		Image:         req.Image,
		CPURequest:    req.CPURequest,
		CPULimit:      req.CPULimit,
		MemoryRequest: req.MemoryRequest,
		MemoryLimit:   req.MemoryLimit,
		MountPaths:    req.MountPaths,
		EnvVars:       req.EnvVars,
	}

	if err := service.ValidateContainerConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	data, err := appService.SaveContainerConfigInTenant(tenantID, config)
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

// ========== 11. 镜像版本查询接口（GET）==========

// ListImageVersions godoc
// @Summary 获取镜像版本列表
// @Description 从Harbor拉取对应应用的镜像版本列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/images/versions [get]
func ListImageVersions(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, _ := strconv.ParseUint(c.Query("app_id"), 10, 64)
	deployService := service.NewDeployConfigService()

	// 根据appID获取应用名称
	appConfig, err := appService.GetAppConfigInTenant(tenantID, uint(appID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    []string{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    deployService.ListImageVersions(appConfig.Name),
	})
}

// ========== 删除配置接口（POST）==========

// DeleteAppConfig godoc
// @Summary 删除应用配置
// @Description 删除应用的基础配置信息
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/config/delete [post]
func DeleteAppConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteAppConfigInTenant(tenantID, req.ID) {
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
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-config/delete [post]
func DeleteBuildConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteBuildConfigInTenant(tenantID, req.ID) {
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
// @Description 删除应用的部署配置（兼容旧接口，删除所有环境配置）
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-config/delete [post]
func DeleteDeployConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteDeployConfigInTenant(tenantID, req.ID) {
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

// DeleteDeployConfigByEnv godoc
// @Summary 删除指定环境的部署配置
// @Description 删除应用指定环境的部署配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteDeployConfigRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/deploy-config/delete-env [post]
func DeleteDeployConfigByEnv(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteDeployConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteDeployConfigByEnvInTenant(tenantID, req.AppID, req.Environment) {
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

// DeleteContainerConfig godoc
// @Summary 删除容器配置
// @Description 删除应用的容器配置（兼容旧接口，删除所有环境配置）
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/container-config/delete [post]
func DeleteContainerConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteContainerConfigInTenant(tenantID, req.ID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "容器配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// DeleteContainerConfigByEnv godoc
// @Summary 删除指定环境的容器配置
// @Description 删除应用指定环境的容器配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteContainerConfigRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/container-config/delete-env [post]
func DeleteContainerConfigByEnv(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteContainerConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteContainerConfigByEnvInTenant(tenantID, req.AppID, req.Environment) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "容器配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
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
			"statusOptions": []gin.H{
				{"label": "运行中", "value": model.StatusRunning},
				{"label": "已下线", "value": model.StatusOffline},
			},
			"instanceTypeOptions": []gin.H{
				{"label": "容器部署", "value": model.InstanceTypeContainer},
				{"label": "原方式部署", "value": model.InstanceTypeNative},
			},
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

// ========== 构建环境版本管理 API ==========

// ListBuildEnvs godoc
// @Summary 获取构建环境版本列表
// @Description 获取所有构建环境版本
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/build-env/list [get]
func ListBuildEnvs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    buildEnvService.ListBuildEnvs(),
	})
}

// SaveBuildEnv godoc
// @Summary 保存构建环境版本
// @Description 新增或编辑JDK/Maven/Golang/Python版本
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.BuildEnv true "构建环境"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-env/save [post]
func SaveBuildEnv(c *gin.Context) {
	var env model.BuildEnv
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	var data model.BuildEnv
	var err error
	if env.ID > 0 {
		data, err = buildEnvService.UpdateBuildEnv(env)
	} else {
		data, err = buildEnvService.CreateBuildEnv(env)
	}
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

// DeleteBuildEnv godoc
// @Summary 删除构建环境版本
// @Description 删除构建环境版本
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/build-env/delete [post]
func DeleteBuildEnv(c *gin.Context) {
	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := buildEnvService.DeleteBuildEnv(req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// ========== 容器配置辅助接口 ==========

// GetEnvPresets godoc
// @Summary 获取环境变量预设
// @Description 获取预设的环境变量模板
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/container-config/env-presets [get]
func GetEnvPresets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    service.GetEnvPresets(),
	})
}

// ========== K8s集群辅助接口 ==========

// ListK8sClusters godoc
// @Summary 获取K8s集群列表
// @Description 获取可用的K8s集群列表
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/k8s/clusters [get]
func ListK8sClusters(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	clusterSvc, err := k8sapi.GetClusterService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群服务失败: " + err.Error(),
		})
		return
	}

	clusters, _, err := clusterSvc.ListInTenant(tenantID, 1, 100, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取K8s集群失败: " + err.Error(),
		})
		return
	}

	data := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		data = append(data, gin.H{
			"id":        cluster.ID,
			"name":      cluster.Name,
			"env":       cluster.Env,
			"status":    cluster.Status,
			"isDefault": cluster.IsDefault,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// ========== 旧版兼容接口（保持向后兼容）==========

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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req service.SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.SaveTemplateInTenant(tenantID, req)
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req service.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.DeployInTenant(tenantID, req)
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req service.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}
	data, err := appService.RollbackInTenant(tenantID, req)
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListTemplatesInTenant(tenantID, c.Query("keyword")),
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListDeploymentsInTenant(tenantID, uint(appID), c.Query("environment"), limit),
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    appService.ListVersionsInTenant(tenantID, uint(appID), limit),
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
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, _ := strconv.ParseUint(c.Query("appId"), 10, 64)
	data, err := appService.QueryTopologyInTenant(tenantID, uint(appID), c.Query("environment"))
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

// GetTechStackConfig godoc
// @Summary 获取技术栈配置
// @Description 获取应用的技术栈配置
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param app_id query int true "应用ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/tech-stack [get]
func GetTechStackConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	appID, err := strconv.ParseUint(c.Query("app_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "app_id参数错误",
		})
		return
	}

	config, err := appService.GetTechStackConfigInTenant(tenantID, uint(appID))
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
// @Router /app/tech-stack/save [post]
func SaveTechStackConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

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

	data, err := appService.SaveTechStackConfigInTenant(tenantID, config)
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

// DeleteTechStackConfig godoc
// @Summary 删除技术栈配置
// @Description 删除应用的技术栈配置
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/tech-stack/delete [post]
func DeleteTechStackConfig(c *gin.Context) {
	tenantID, ok := requireTenantID(c)
	if !ok {
		return
	}

	var req model.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !appService.DeleteTechStackConfigInTenant(tenantID, req.ID) {
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

// ========== 枚举管理 API ==========

// ListAllEnums godoc
// @Summary 获取枚举列表
// @Description 获取所有枚举或按类型筛选，支持管理页面查看所有（包括禁用的）
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Param enum_type query string false "枚举类型筛选"
// @Param include_disabled query bool false "是否包含禁用的枚举"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/enums/list [get]
func ListAllEnums(c *gin.Context) {
	enumType := c.Query("enum_type")
	includeDisabled := c.Query("include_disabled") == "true"

	var enums []model.Enum
	if includeDisabled {
		enums = enumService.ListAllEnums(enumType)
	} else {
		enums = enumService.ListEnums(enumType)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    enums,
	})
}

// GetEnumTypes godoc
// @Summary 获取枚举类型列表
// @Description 获取所有可用的枚举类型
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/enums/types [get]
func GetEnumTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    enumService.GetEnumTypes(),
	})
}

// SaveEnum godoc
// @Summary 保存枚举
// @Description 新增或修改枚举值
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.EnumRequest true "枚举请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/enums/save [post]
func SaveEnum(c *gin.Context) {
	var req model.EnumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	data, err := enumService.SaveEnum(req)
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

// DeleteEnum godoc
// @Summary 删除枚举
// @Description 删除枚举值（被使用的枚举无法删除）
// @Tags 应用管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.DeleteEnumRequest true "删除请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /app/enums/delete [post]
func DeleteEnum(c *gin.Context) {
	var req model.DeleteEnumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := enumService.DeleteEnum(req.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetEnumsGrouped godoc
// @Summary 获取分组枚举
// @Description 获取按类型分组的枚举，用于前端下拉框
// @Tags 应用管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /app/enums/grouped [get]
func GetEnumsGrouped(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    enumService.GetEnumsByType(),
	})
}
