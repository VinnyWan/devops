package routers

import (
	"devops/controller"
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes 设置用户管理路由
func SetupUserRoutes(r *gin.RouterGroup) {
	userCtrl := controller.NewUserController()

	// 需要JWT认证
	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	{
		// 当前用户信息
		auth.GET("/user/info", userCtrl.GetInfo)

		// 用户管理CRUD - 按ID
		auth.GET("/users", userCtrl.GetList)
		auth.POST("/users", userCtrl.Create)
		auth.GET("/users/:id", userCtrl.GetByID)
		auth.PUT("/users/:id", userCtrl.Update)
		auth.DELETE("/users/:id", userCtrl.Delete)
		auth.POST("/users/:id/roles", userCtrl.AssignRoles)

		// 用户管理 - 按用户名
		auth.GET("/users/username/:username", userCtrl.GetByUsername)
		auth.PUT("/users/username/:username", userCtrl.UpdateByUsername)
		auth.DELETE("/users/username/:username", userCtrl.DeleteByUsername)
	}
}
