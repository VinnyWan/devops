package v1

import (
	"devops-platform/internal/middleware"
	appAPI "devops-platform/internal/modules/app/api"

	"github.com/gin-gonic/gin"
)

func registerApp(r *gin.RouterGroup) {
	g := r.Group("/app")
	listPermission := middleware.RequirePermission("app", "list")
	createPermission := middleware.RequirePermission("app", "create")
	deletePermission := middleware.RequirePermission("app", "delete")

	{
		// ========== 菜单选项 ==========
		g.GET("/menu/options", listPermission, appAPI.GetAppManagementMenuOptions)

		// ========== 1. 应用列表 + 搜索 + 筛选接口（GET）==========
		g.GET("/list", listPermission, appAPI.ListApps)

		// ========== 2. 单应用配置详情查询接口（GET）==========
		g.GET("/config", listPermission, appAPI.GetAppConfig)

		// ========== 3. 应用配置新增/修改接口（POST）==========
		g.POST("/config/save", createPermission, middleware.SetAuditOperation("保存应用配置"), appAPI.SaveAppConfig)

		// ========== 4. 应用状态切换接口（POST）==========
		g.POST("/status/toggle", createPermission, middleware.SetAuditOperation("切换应用状态"), appAPI.ToggleAppStatus)

		// ========== 5. 构建配置查询接口（GET）==========
		g.GET("/build-config", listPermission, appAPI.GetBuildConfig)

		// ========== 6. 构建配置保存/修改接口（POST）==========
		g.POST("/build-config/save", createPermission, middleware.SetAuditOperation("保存构建配置"), appAPI.SaveBuildConfig)

		// ========== 7. 部署配置查询接口（GET）==========
		g.GET("/deploy-config", listPermission, appAPI.GetDeployConfig)
		g.GET("/deploy-configs", listPermission, appAPI.ListDeployConfigs)

		// ========== 8. 部署配置保存/修改接口（POST）==========
		g.POST("/deploy-config/save", createPermission, middleware.SetAuditOperation("保存部署配置"), appAPI.SaveDeployConfig)
		g.POST("/deploy-config/delete-env", deletePermission, middleware.SetAuditOperation("删除指定环境部署配置"), appAPI.DeleteDeployConfigByEnv)

		// ========== 9. 容器配置查询接口（GET）==========
		g.GET("/container-config", listPermission, appAPI.GetContainerConfig)
		g.GET("/container-configs", listPermission, appAPI.ListContainerConfigs)

		// ========== 10. 容器配置保存/修改接口（POST）==========
		g.POST("/container-config/save", createPermission, middleware.SetAuditOperation("保存容器配置"), appAPI.SaveContainerConfig)
		g.POST("/container-config/delete-env", deletePermission, middleware.SetAuditOperation("删除指定环境容器配置"), appAPI.DeleteContainerConfigByEnv)

		// ========== 11. 镜像版本查询接口（GET）==========
		g.GET("/images/versions", listPermission, appAPI.ListImageVersions)

		// ========== 删除配置接口（POST）==========
		g.POST("/config/delete", deletePermission, middleware.SetAuditOperation("删除应用配置"), appAPI.DeleteAppConfig)
		g.POST("/build-config/delete", deletePermission, middleware.SetAuditOperation("删除构建配置"), appAPI.DeleteBuildConfig)
		g.POST("/deploy-config/delete", deletePermission, middleware.SetAuditOperation("删除部署配置"), appAPI.DeleteDeployConfig)
		g.POST("/container-config/delete", deletePermission, middleware.SetAuditOperation("删除容器配置"), appAPI.DeleteContainerConfig)

		// ========== 枚举值接口（GET）==========
		g.GET("/enums", listPermission, appAPI.GetEnumOptions)

		// ========== 枚举管理接口 ==========
		g.GET("/enums/list", listPermission, appAPI.ListAllEnums)
		g.GET("/enums/types", listPermission, appAPI.GetEnumTypes)
		g.GET("/enums/grouped", listPermission, appAPI.GetEnumsGrouped)
		g.POST("/enums/save", createPermission, middleware.SetAuditOperation("保存枚举值"), appAPI.SaveEnum)
		g.POST("/enums/delete", deletePermission, middleware.SetAuditOperation("删除枚举值"), appAPI.DeleteEnum)

		// ========== 构建环境版本管理（统一POST）==========
		g.GET("/build-env/list", listPermission, appAPI.ListBuildEnvs)
		g.POST("/build-env/save", createPermission, middleware.SetAuditOperation("保存构建环境"), appAPI.SaveBuildEnv)
		g.POST("/build-env/delete", deletePermission, middleware.SetAuditOperation("删除构建环境"), appAPI.DeleteBuildEnv)

		// ========== 容器配置辅助接口 ==========
		g.GET("/container-config/env-presets", listPermission, appAPI.GetEnvPresets)

		// ========== K8s集群辅助接口 ==========
		g.GET("/k8s/clusters", listPermission, appAPI.ListK8sClusters)

		// ========== 技术栈配置 ==========
		g.GET("/tech-stack", listPermission, appAPI.GetTechStackConfig)
		g.POST("/tech-stack/save", createPermission, middleware.SetAuditOperation("保存技术栈配置"), appAPI.SaveTechStackConfig)
		g.POST("/tech-stack/delete", deletePermission, middleware.SetAuditOperation("删除技术栈配置"), appAPI.DeleteTechStackConfig)

		// ========== 旧版兼容接口 ==========
		g.GET("/template/list", listPermission, appAPI.ListTemplates)
		g.POST("/template/save", createPermission, middleware.SetAuditOperation("管理应用模板"), appAPI.CreateApp)
		g.POST("/deploy", createPermission, middleware.SetAuditOperation("多环境部署应用"), appAPI.UpdateApp)
		g.GET("/deployment/list", listPermission, appAPI.ListDeployments)
		g.GET("/version/list", listPermission, appAPI.ListVersions)
		g.POST("/rollback", deletePermission, middleware.SetAuditOperation("应用版本回滚"), appAPI.DeleteApp)
		g.GET("/topology", listPermission, appAPI.QueryTopology)
	}
}
