package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerPlatformRoutes(v1 *gin.RouterGroup) {
	platform := v1.Group("/platform", authMiddlewares()...)
	{
		tenants := platform.Group("/tenants")
		{
			tenants.GET("", middleware.RequirePermission("tenant", "list"), api.ListTenants)
			tenants.GET("/:id", middleware.RequirePermission("tenant", "list"), api.GetTenantByID)
			tenants.POST("", middleware.RequirePermission("tenant", "create"), middleware.SetAuditOperation("创建租户"), api.CreateTenant)
			tenants.PUT("/:id", middleware.RequirePermission("tenant", "update"), middleware.SetAuditOperation("更新租户"), api.UpdateTenantREST)
			tenants.DELETE("/:id", middleware.RequirePermission("tenant", "delete"), middleware.SetAuditOperation("停用租户"), api.DisableTenantREST)
		}
	}
}
