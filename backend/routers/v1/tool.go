package v1

import (
	toolAPI "devops-platform/internal/modules/tool/api"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerToolRoutes(r *gin.RouterGroup) {
	toolAPI.InstallToolRoutes(r)

	g := r.Group("/tools")
	queryPermission := middleware.RequirePermission("tool", "list")
	updatePermission := middleware.RequirePermission("tool", "update")

	// Templates
	g.GET("/templates", queryPermission, toolAPI.ListTemplates)
	g.GET("/templates/:id", queryPermission, toolAPI.GetTemplate)
	g.POST("/templates", updatePermission,
		middleware.SetAuditOperation("创建工具模板"),
		toolAPI.SaveTemplate)
	g.PUT("/templates/:id", updatePermission, toolAPI.SaveTemplate)
	g.DELETE("/templates/:id", updatePermission,
		middleware.SetAuditOperation("删除工具模板"),
		toolAPI.DeleteTemplate)

	// Template Versions
	g.GET("/templates/:id/versions", queryPermission, toolAPI.ListTemplateVersions)
	g.POST("/templates/:id/versions", updatePermission,
		middleware.SetAuditOperation("创建模板版本"),
		toolAPI.SaveTemplateVersion)
	g.DELETE("/templates/:id/versions/:versionId", updatePermission,
		middleware.SetAuditOperation("删除模板版本"),
		toolAPI.DeleteTemplateVersion)
}
