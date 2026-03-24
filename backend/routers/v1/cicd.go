package v1

import (
	"devops-platform/internal/middleware"
	cicdAPI "devops-platform/internal/modules/cicd/api"

	"github.com/gin-gonic/gin"
)

func registerCICD(r *gin.RouterGroup) {
	g := r.Group("/cicd")
	listPermission := middleware.RequirePermission("cicd", "list")
	triggerPermission := middleware.RequirePermission("cicd", "trigger")
	templatePermission := middleware.RequirePermission("cicd", "template")
	updatePermission := middleware.RequirePermission("cicd", "update")
	{
		g.GET("/list", listPermission, cicdAPI.ListPipelines)
		g.GET("/status", listPermission, cicdAPI.ListPipelineStatus)
		g.GET("/logs", listPermission, cicdAPI.ListPipelineLogs)
		g.GET("/templates", listPermission, cicdAPI.ListPipelineTemplates)
		g.POST("/template/save", templatePermission, middleware.SetAuditOperation("保存流水线模板"), cicdAPI.SavePipelineTemplate)
		g.GET("/orchestration/preview", listPermission, cicdAPI.PreviewPipelineOrchestration)
		g.POST("/trigger", triggerPermission, middleware.SetAuditOperation("触发流水线"), cicdAPI.TriggerPipeline)
		g.GET("/runs", listPermission, cicdAPI.ListPipelineRuns)
		g.GET("/config", listPermission, cicdAPI.GetJenkinsConfig)
		g.POST("/config/upsert", updatePermission, middleware.SetAuditOperation("Jenkins 配置更新"), cicdAPI.SaveJenkinsConfig)
	}
}
