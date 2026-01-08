package routers

import (
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupDepartmentRoutes 设置部门管理路由
func SetupDepartmentRoutes(r *gin.RouterGroup) {
	// TODO: 创建部门控制器后取消注释
	// deptCtrl := controller.NewDepartmentController()

	// 需要JWT认证
	auth := r.Group("/departments")
	auth.Use(middleware.JWTAuth())
	{
		// auth.GET("", deptCtrl.GetList)
		// auth.GET("/tree", deptCtrl.GetTreeList)
		// auth.POST("", deptCtrl.Create)
		// auth.GET("/:id", deptCtrl.GetByID)
		// auth.PUT("/:id", deptCtrl.Update)
		// auth.DELETE("/:id", deptCtrl.Delete)
	}
}
