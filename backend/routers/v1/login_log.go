package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerLoginLog(r *gin.RouterGroup) {
	g := r.Group("/login-log")
	permission := middleware.RequirePermission("login-log", "list")

	g.GET("/list", permission, api.ListLoginLogs)
}
