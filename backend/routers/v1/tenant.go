package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerTenant(r *gin.RouterGroup) {
	g := r.Group("/tenant")
	{
		g.GET("/list", middleware.RequirePermission("tenant", "list"), api.ListTenants)
		g.GET("/detail", middleware.RequirePermission("tenant", "list"), api.GetTenantDetail)
		g.POST("/create", middleware.RequirePermission("tenant", "create"), middleware.SetAuditOperation("创建租户"), api.CreateTenant)
		g.POST("/update", middleware.RequirePermission("tenant", "update"), middleware.SetAuditOperation("更新租户"), api.UpdateTenant)
		g.POST("/disable", middleware.RequirePermission("tenant", "delete"), middleware.SetAuditOperation("停用租户"), api.DisableTenant)
	}
}
