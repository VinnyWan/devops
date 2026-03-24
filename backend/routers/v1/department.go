package v1

import (
	"devops-platform/internal/middleware"
	"devops-platform/internal/modules/user/api"

	"github.com/gin-gonic/gin"
)

func registerDepartment(r *gin.RouterGroup) {
	g := r.Group("/department")
	{
		g.GET("/list", middleware.RequirePermission("department", "list"), api.ListDepartments)
		g.POST("/create", middleware.RequirePermission("department", "create"), middleware.SetAuditOperation("创建部门"), api.CreateDepartment)
		g.POST("/update", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("更新部门"), api.UpdateDepartment)
		g.POST("/delete", middleware.RequirePermission("department", "delete"), middleware.SetAuditOperation("删除部门"), api.DeleteDepartment)
		g.POST("/assign-roles", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("部门分配角色"), api.AssignDeptRoles)

		users := g.Group("/users")
		{
			users.GET("/list", middleware.RequirePermission("department", "list"), api.ListDepartmentUsers)
			users.POST("/create", middleware.RequirePermission("department", "create"), middleware.SetAuditOperation("创建部门用户"), api.CreateDepartmentUser)
			users.POST("/update", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("更新部门用户"), api.UpdateDepartmentUser)
			users.POST("/delete", middleware.RequirePermission("department", "delete"), middleware.SetAuditOperation("删除部门用户"), api.DeleteDepartmentUser)
			users.POST("/transfer", middleware.RequirePermission("department", "update"), middleware.SetAuditOperation("用户切换部门"), api.TransferDepartmentUser)
		}
	}
}
