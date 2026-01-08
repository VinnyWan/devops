package routers

import (
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMenuRoutes 设置菜单管理路由
func SetupMenuRoutes(r *gin.RouterGroup) {
	// TODO: 创建菜单控制器后取消注释
	// menuCtrl := controller.NewMenuController()

	// 需要JWT认证
	auth := r.Group("/menus")
	auth.Use(middleware.JWTAuth())
	{
		// auth.GET("", menuCtrl.GetList)
		// auth.GET("/tree", menuCtrl.GetTreeList)
		// auth.POST("", menuCtrl.Create)
		// auth.GET("/:id", menuCtrl.GetByID)
		// auth.PUT("/:id", menuCtrl.Update)
		// auth.DELETE("/:id", menuCtrl.Delete)
	}
}
