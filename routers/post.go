package routers

import (
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPostRoutes 设置岗位管理路由
func SetupPostRoutes(r *gin.RouterGroup) {
	// TODO: 创建岗位控制器后取消注释
	// postCtrl := controller.NewPostController()

	// 需要JWT认证
	auth := r.Group("/posts")
	auth.Use(middleware.JWTAuth())
	{
		// auth.GET("", postCtrl.GetList)
		// auth.POST("", postCtrl.Create)
		// auth.GET("/:id", postCtrl.GetByID)
		// auth.PUT("/:id", postCtrl.Update)
		// auth.DELETE("/:id", postCtrl.Delete)
	}
}
