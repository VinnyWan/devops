package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerUser(r *gin.RouterGroup) {
	// 用户自身接口 (仅需登录)
	r.GET("/user/info", api.GetUserInfo)
	r.POST("/user/change-password", middleware.SetAuditOperation("用户修改密码"), api.ChangePassword)
	r.POST("/user/permissions", api.GetUserPermissions)
	r.GET("/user/all-permissions", api.GetAllPermissions)

	// 用户管理接口 (需要权限)
	g := r.Group("/user")
	{
		// 用户列表与详情 (查看权限)
		g.GET("/list", middleware.RequirePermission("user", "list"), api.List)
		g.GET("/detail", middleware.RequirePermission("user", "list"), api.GetDetail)

		// 用户更新与删除 (管理权限)
		g.POST("/update", middleware.RequirePermission("user", "update"), middleware.SetAuditOperation("更新用户"), api.Update)
		g.POST("/delete", middleware.RequirePermission("user", "delete"), middleware.SetAuditOperation("删除用户"), api.Delete)

		// 敏感操作
		g.POST("/reset-password", middleware.RequirePermission("user", "reset_password"), middleware.SetAuditOperation("重置用户密码"), api.ResetPassword)
		g.POST("/assign-roles", middleware.RequirePermission("user", "assign_roles"), middleware.SetAuditOperation("用户分配角色"), api.AssignRoles)
		g.POST("/lock", middleware.RequirePermission("user", "lock"), middleware.SetAuditOperation("锁定用户"), api.LockUser)
		g.POST("/unlock", middleware.RequirePermission("user", "unlock"), middleware.SetAuditOperation("解锁用户"), api.UnlockUser)
	}
}
