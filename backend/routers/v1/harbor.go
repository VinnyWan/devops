package v1

import (
	"devops-platform/internal/middleware"
	harborAPI "devops-platform/internal/modules/harbor/api"

	"github.com/gin-gonic/gin"
)

func registerHarbor(r *gin.RouterGroup) {
	g := r.Group("/harbor")
	listPermission := middleware.RequirePermission("harbor", "list")
	updatePermission := middleware.RequirePermission("harbor", "update")
	{
		g.GET("/list", listPermission, harborAPI.ListHarborProjects)
		g.GET("/images", listPermission, harborAPI.ListHarborImages)
		g.GET("/config", listPermission, harborAPI.GetHarborConfig)
		g.POST("/config/upsert", updatePermission, middleware.SetAuditOperation("Harbor 配置更新"), harborAPI.SaveHarborConfig)
	}
}
