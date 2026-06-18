package v1

import (
	cicdAPI "devops-platform/internal/modules/cicd/api"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerCICD(r *gin.RouterGroup) {
	g := r.Group("/cicd")
	queryPermission := middleware.RequirePermission("cicd", "list")
	updatePermission := middleware.RequirePermission("cicd", "update")

	// Jenkins config
	g.GET("/jenkins", queryPermission, cicdAPI.ListJenkinsConfigs)
	g.POST("/jenkins", updatePermission,
		middleware.SetAuditOperation("Jenkins 配置保存"),
		cicdAPI.SaveJenkinsConfig)
	g.PUT("/jenkins/:id", updatePermission,
		middleware.SetAuditOperation("Jenkins 配置更新"),
		cicdAPI.SaveJenkinsConfig)
	g.DELETE("/jenkins/:id", updatePermission,
		middleware.SetAuditOperation("Jenkins 配置删除"),
		cicdAPI.DeleteJenkinsConfig)
	g.POST("/jenkins/test", queryPermission, cicdAPI.TestJenkinsConnection)

	// Jobs & builds
	g.GET("/jenkins/:configId/jobs", queryPermission, cicdAPI.ListJobs)
	g.POST("/jenkins/:configId/build", updatePermission,
		middleware.SetAuditOperation("触发 Jenkins 构建"),
		cicdAPI.TriggerBuild)
	g.GET("/jenkins/:configId/builds", queryPermission, cicdAPI.ListBuilds)
	g.GET("/jenkins/:configId/build-log", queryPermission, cicdAPI.GetBuildLog)

	// Pipelines
	g.GET("/pipelines", queryPermission, cicdAPI.ListPipelines)
	g.POST("/pipelines", updatePermission,
		middleware.SetAuditOperation("创建流水线"),
		cicdAPI.SavePipeline)
	g.PUT("/pipelines/:id", updatePermission, cicdAPI.SavePipeline)
	g.DELETE("/pipelines/:id", updatePermission,
		middleware.SetAuditOperation("删除流水线"),
		cicdAPI.DeletePipeline)
}
