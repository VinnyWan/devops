package v1

import (
	"devops-platform/internal/middleware"
	logAPI "devops-platform/internal/modules/log/api"

	"github.com/gin-gonic/gin"
)

func registerLog(r *gin.RouterGroup) {
	g := r.Group("/log")
	listPermission := middleware.RequirePermission("log", "list")
	{
		g.GET("/search", listPermission, logAPI.SearchLogs)
	}
}
