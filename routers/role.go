package routers

import (
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoleRoutes 设置角色管理路由
func SetupRoleRoutes(r *gin.RouterGroup) {
	// TODO: 创建角色控制器后取消注释
	// roleCtrl := controller.NewRoleController()

	// 需要JWT认证
	auth := r.Group("/roles")
	auth.Use(middleware.JWTAuth())
	{
		// auth.GET("", roleCtrl.GetList)
		// auth.POST("", roleCtrl.Create)
		// auth.GET("/:id", roleCtrl.GetByID)
		// auth.PUT("/:id", roleCtrl.Update)
		// auth.DELETE("/:id", roleCtrl.Delete)
		// auth.POST("/:id/menus", roleCtrl.AssignMenus)
	}
}
