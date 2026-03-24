package v1

import (
	"devops-platform/internal/middleware"
	monitorAPI "devops-platform/internal/modules/monitor/api"

	"github.com/gin-gonic/gin"
)

func registerMonitor(r *gin.RouterGroup) {
	g := r.Group("/monitor")
	queryPermission := middleware.RequirePermission("monitor", "list")
	updatePermission := middleware.RequirePermission("monitor", "update")
	{
		g.GET("/query", queryPermission, monitorAPI.QueryMonitors)
		g.GET("/config", queryPermission, monitorAPI.GetPrometheusConfig)
		g.POST("/config/upsert", updatePermission, middleware.SetAuditOperation("Prometheus 配置更新"), monitorAPI.SavePrometheusConfig)
	}
}
