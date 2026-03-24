package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerAudit(r *gin.RouterGroup) {
	g := r.Group("/audit")
	permission := middleware.RequirePermission("audit", "list")

	g.GET("/list", permission, api.ListAuditLogs)
	g.GET("/export", permission, api.ExportAuditLogs)
	g.POST("/cleanup", permission, api.CleanupExpiredAuditLogs)
}
