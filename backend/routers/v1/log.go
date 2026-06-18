package v1

import (
	logAPI "devops-platform/internal/modules/log/api"
	"devops-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerLog(r *gin.RouterGroup) {
	g := r.Group("/log")
	queryPermission := middleware.RequirePermission("log", "list")
	updatePermission := middleware.RequirePermission("log", "update")

	// Source management
	g.GET("/sources", queryPermission, logAPI.ListLogSources)
	g.POST("/sources", updatePermission,
		middleware.SetAuditOperation("日志源保存"),
		logAPI.SaveLogSource)
	g.PUT("/sources/:id", updatePermission, logAPI.SaveLogSource)
	g.DELETE("/sources/:id", updatePermission,
		middleware.SetAuditOperation("日志源删除"),
		logAPI.DeleteLogSource)
	g.POST("/sources/:id/test", queryPermission, logAPI.TestLogSourceConnection)

	// Search & Export
	g.POST("/search", queryPermission, logAPI.SearchLogs)
	g.POST("/export", queryPermission, logAPI.ExportLogs)
}
