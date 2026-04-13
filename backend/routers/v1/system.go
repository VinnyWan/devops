package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerSystemRoutes(v1 *gin.RouterGroup) {
	system := v1.Group("/system", authMiddlewares()...)
	{
		registerSystemUsers(system)
		registerSystemRoles(system)
		registerSystemPermissions(system)
		registerSystemDepartments(system)
	}
}

func registerSystemUsers(system *gin.RouterGroup) {
	users := system.Group("/users")
	{
		users.GET("", middleware.RequirePermission("user", "list"), api.List)
		users.GET("/:id", middleware.RequirePermission("user", "list"), api.GetUserByID)
		users.POST("", middleware.RequirePermission("user", "create"), middleware.SetAuditOperation("创建用户"), api.CreateUserREST)
		users.PUT("/:id", middleware.RequirePermission("user", "update"), middleware.SetAuditOperation("更新用户"), api.UpdateUserREST)
		users.DELETE("/:id", middleware.RequirePermission("user", "delete"), middleware.SetAuditOperation("删除用户"), api.DeleteUserREST)
		users.PUT("/:id/roles", middleware.RequirePermission("user", "assign_roles"), middleware.SetAuditOperation("用户分配角色"), api.AssignRolesREST)
		users.PUT("/:id/password", api.ChangePasswordByID)
		users.PUT("/:id/reset-password", middleware.RequirePermission("user", "reset_password"), middleware.SetAuditOperation("重置用户密码"), api.ResetPasswordREST)
		users.PUT("/:id/lock", middleware.RequirePermission("user", "lock"), middleware.SetAuditOperation("锁定用户"), api.LockUserREST)
		users.PUT("/:id/unlock", middleware.RequirePermission("user", "unlock"), middleware.SetAuditOperation("解锁用户"), api.UnlockUserREST)
	}
}

func registerSystemRoles(system *gin.RouterGroup) {
	roles := system.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission("role", "list"), api.ListRoles)
		roles.POST("", middleware.RequirePermission("role", "create"), middleware.SetAuditOperation("创建角色"), api.CreateRole)
		roles.GET("/:id", middleware.RequirePermission("role", "list"), api.GetRoleByID)
		roles.PUT("/:id", middleware.RequirePermission("role", "update"), middleware.SetAuditOperation("更新角色"), api.UpdateRoleREST)
		roles.DELETE("/:id", middleware.RequirePermission("role", "delete"), middleware.SetAuditOperation("删除角色"), api.DeleteRoleREST)
		roles.PUT("/:id/permissions", middleware.RequirePermission("role", "update"), middleware.SetAuditOperation("角色分配权限"), api.AssignPermissionsREST)
		roles.PUT("/:id/users", middleware.RequirePermission("role", "update"), middleware.SetAuditOperation("角色关联用户"), api.AssignRoleUsersREST)
		roles.PUT("/:id/departments", middleware.RequirePermission("role", "update"), middleware.SetAuditOperation("角色关联部门"), api.AssignRoleDepartmentsREST)
		roles.GET("/:id/users", middleware.RequirePermission("role", "list"), api.GetRoleUsersREST)
		roles.GET("/:id/departments", middleware.RequirePermission("role", "list"), api.GetRoleDepartmentsREST)
	}
}

func registerSystemPermissions(system *gin.RouterGroup) {
	permissions := system.Group("/permissions")
	{
		permissions.GET("", middleware.RequirePermission("permission", "list"), api.ListPermissions)
		permissions.POST("", middleware.RequirePermission("permission", "create"), middleware.SetAuditOperation("创建权限"), api.CreatePermission)
		permissions.GET("/all", middleware.RequirePermission("permission", "list"), api.ListAllPermissions)
		permissions.GET("/:id", middleware.RequirePermission("permission", "list"), api.GetPermissionByID)
		permissions.PUT("/:id", middleware.RequirePermission("permission", "update"), middleware.SetAuditOperation("更新权限"), api.UpdatePermissionREST)
		permissions.DELETE("/:id", middleware.RequirePermission("permission", "delete"), middleware.SetAuditOperation("删除权限"), api.DeletePermissionREST)
	}
}

func registerSystemDepartments(system *gin.RouterGroup) {
	departments := system.Group("/departments")
	{
		departments.GET("/tree", middleware.RequirePermission("department", "list"), api.ListDepartments)
		departments.POST("", middleware.RequirePermission("department", "create"), middleware.SetAuditOperation("创建部门"), api.CreateDepartment)
		departments.PUT("/:id", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("更新部门"), api.UpdateDepartmentREST)
		departments.DELETE("/:id", middleware.RequirePermission("department", "delete"), middleware.SetAuditOperation("删除部门"), api.DeleteDepartmentREST)
		departments.PUT("/:id/roles", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("部门分配角色"), api.AssignDeptRolesREST)

		// 部门用户子资源
		departments.GET("/:id/users", middleware.RequirePermission("department", "list"), api.ListDepartmentUsersREST)
		departments.POST("/:id/users", middleware.RequirePermission("department", "create"), middleware.SetAuditOperation("创建部门用户"), api.CreateDepartmentUserREST)
		departments.PUT("/:id/users/:uid", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("更新部门用户"), api.UpdateDepartmentUserREST)
		departments.DELETE("/:id/users/:uid", middleware.RequirePermission("department", "delete"), middleware.SetAuditOperation("删除部门用户"), api.DeleteDepartmentUserREST)
	}
}
