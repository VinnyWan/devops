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
		g.GET("/list", listPermission, appAPI.ListApps)
		g.GET("/template/list", listPermission, appAPI.ListTemplates)
		g.POST("/template/save", createPermission, middleware.SetAuditOperation("管理应用模板"), appAPI.CreateApp)
		g.POST("/deploy", updatePermission, middleware.SetAuditOperation("多环境部署应用"), appAPI.UpdateApp)
		g.GET("/deployment/list", listPermission, appAPI.ListDeployments)
		g.GET("/version/list", listPermission, appAPI.ListVersions)
		g.POST("/rollback", deletePermission, middleware.SetAuditOperation("应用版本回滚"), appAPI.DeleteApp)
		g.GET("/topology", listPermission, appAPI.QueryTopology)
	}
}
