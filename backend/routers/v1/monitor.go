package v1

import (
	monitorAPI "devops-platform/internal/modules/monitor/api"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerMonitor(r *gin.RouterGroup) {
	g := r.Group("/monitor")
	queryPermission := middleware.RequirePermission("monitor", "list")
	updatePermission := middleware.RequirePermission("monitor", "update")

	// Prometheus config management
	g.GET("/prometheus", queryPermission, monitorAPI.ListPrometheusConfigs)
	g.GET("/prometheus/:id", queryPermission, monitorAPI.GetPrometheusConfig)
	g.POST("/prometheus", updatePermission,
		middleware.SetAuditOperation("Prometheus 配置保存"),
		monitorAPI.SavePrometheusConfig)
	g.PUT("/prometheus/:id", updatePermission,
		middleware.SetAuditOperation("Prometheus 配置更新"),
		monitorAPI.SavePrometheusConfig)
	g.DELETE("/prometheus/:id", updatePermission,
		middleware.SetAuditOperation("Prometheus 配置删除"),
		monitorAPI.DeletePrometheusConfig)
	g.POST("/prometheus/test", queryPermission, monitorAPI.TestPrometheusConnection)

	// Host metrics
	g.GET("/host/metrics", queryPermission, monitorAPI.QueryHostMetrics)
	g.GET("/host/ports", queryPermission, monitorAPI.QueryPortStatus)

	// Agent management
	g.GET("/agent/status", queryPermission, monitorAPI.QueryAgentStatus)
}
