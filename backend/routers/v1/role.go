package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerRole(r *gin.RouterGroup) {
	// 角色管理
	roleGroup := r.Group("/role")
	roleListPermission := middleware.RequirePermission("role", "list")
	roleCreatePermission := middleware.RequirePermission("role", "create")
	roleUpdatePermission := middleware.RequirePermission("role", "update")
	roleDeletePermission := middleware.RequirePermission("role", "delete")
	{
		roleGroup.POST("/create", roleCreatePermission, middleware.SetAuditOperation("创建角色"), api.CreateRole)
		roleGroup.GET("/list", roleListPermission, api.ListRoles)
		roleGroup.GET("/detail", roleListPermission, api.GetRoleDetail)
		roleGroup.POST("/update", roleUpdatePermission, middleware.SetAuditOperation("更新角色"), api.UpdateRole)
		roleGroup.POST("/delete", roleDeletePermission, middleware.SetAuditOperation("删除角色"), api.DeleteRole)
		roleGroup.POST("/assign-permissions", roleUpdatePermission, middleware.SetAuditOperation("角色分配权限"), api.AssignPermissions)
		roleGroup.POST("/assign-users", roleUpdatePermission, middleware.SetAuditOperation("角色关联用户"), api.AssignRoleUsers)
		roleGroup.POST("/assign-departments", roleUpdatePermission, middleware.SetAuditOperation("角色关联部门"), api.AssignRoleDepartments)
		roleGroup.GET("/users", roleListPermission, api.GetRoleUsers)
		roleGroup.GET("/departments", roleListPermission, api.GetRoleDepartments)
	}

	// 权限管理
	permGroup := r.Group("/permission")
	permListPermission := middleware.RequirePermission("permission", "list")
	permCreatePermission := middleware.RequirePermission("permission", "create")
	permUpdatePermission := middleware.RequirePermission("permission", "update")
	permDeletePermission := middleware.RequirePermission("permission", "delete")
	{
		permGroup.POST("/create", permCreatePermission, middleware.SetAuditOperation("创建权限"), api.CreatePermission)
		permGroup.GET("/list", permListPermission, api.ListPermissions)
		permGroup.GET("/all", permListPermission, api.ListAllPermissions)
		permGroup.GET("/detail", permListPermission, api.GetPermissionDetail)
		permGroup.POST("/update", permUpdatePermission, middleware.SetAuditOperation("更新权限"), api.UpdatePermission)
		permGroup.POST("/delete", permDeletePermission, middleware.SetAuditOperation("删除权限"), api.DeletePermission)
	}
}
