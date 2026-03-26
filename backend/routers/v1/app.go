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
	updatePermission := middleware.RequirePermission("app", "update")
	deletePermission := middleware.RequirePermission("app", "delete")
	{
		// 原有的路由
		g.GET("/list", listPermission, appAPI.ListApps)
		g.GET("/template/list", listPermission, appAPI.ListTemplates)
		g.POST("/template/save", createPermission, middleware.SetAuditOperation("管理应用模板"), appAPI.CreateApp)
		g.POST("/deploy", updatePermission, middleware.SetAuditOperation("多环境部署应用"), appAPI.UpdateApp)
		g.GET("/deployment/list", listPermission, appAPI.ListDeployments)
		g.GET("/version/list", listPermission, appAPI.ListVersions)
		g.POST("/rollback", deletePermission, middleware.SetAuditOperation("应用版本回滚"), appAPI.DeleteApp)
		g.GET("/topology", listPermission, appAPI.QueryTopology)

		// ========== 枚举值路由 ==========
		g.GET("/enums", listPermission, appAPI.GetEnumOptions)

		// ========== 应用配置 CRUD ==========
		g.GET("/:appId/config", listPermission, appAPI.GetAppConfig)
		g.POST("/config", createPermission, middleware.SetAuditOperation("保存应用配置"), appAPI.SaveAppConfig)
		g.DELETE("/:appId/config", deletePermission, middleware.SetAuditOperation("删除应用配置"), appAPI.DeleteAppConfig)

		// ========== 构建配置 CRUD ==========
		g.GET("/:appId/build-config", listPermission, appAPI.GetBuildConfig)
		g.POST("/build-config", createPermission, middleware.SetAuditOperation("保存构建配置"), appAPI.SaveBuildConfig)
		g.DELETE("/:appId/build-config", deletePermission, middleware.SetAuditOperation("删除构建配置"), appAPI.DeleteBuildConfig)

		// ========== 部署配置 CRUD ==========
		g.GET("/:appId/deploy-config", listPermission, appAPI.GetDeployConfig)
		g.POST("/deploy-config", createPermission, middleware.SetAuditOperation("保存部署配置"), appAPI.SaveDeployConfig)
		g.DELETE("/:appId/deploy-config", deletePermission, middleware.SetAuditOperation("删除部署配置"), appAPI.DeleteDeployConfig)

		// ========== 技术栈配置 CRUD ==========
		g.GET("/:appId/tech-stack", listPermission, appAPI.GetTechStackConfig)
		g.POST("/tech-stack", createPermission, middleware.SetAuditOperation("保存技术栈配置"), appAPI.SaveTechStackConfig)
		g.DELETE("/:appId/tech-stack", deletePermission, middleware.SetAuditOperation("删除技术栈配置"), appAPI.DeleteTechStackConfig)
	}
}
