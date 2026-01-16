package user

import (
	userctrl "devops/controller/user"
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes 设置用户管理路由
func SetupUserRoutes(r *gin.RouterGroup) {
	userCtrl := userctrl.NewUserController()

	// 需要JWT认证
	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	{
		// 当前用户信息
		auth.GET("/user/info", userCtrl.GetInfo)

		// 用户管理CRUD
		auth.GET("/users", userCtrl.GetList)
		auth.POST("/user/create", userCtrl.Create)
		auth.GET("/user/detail", userCtrl.GetByID)
		auth.POST("/user/update", userCtrl.Update)        // 改为POST
		auth.POST("/user/delete", userCtrl.Delete)        // 改为POST
		auth.POST("/user/roles", userCtrl.AssignRoles)

		// 用户管理 - 按用户名
		auth.GET("/user/by-username", userCtrl.GetByUsername)
		auth.POST("/user/update-by-username", userCtrl.UpdateByUsername)  // 改为POST
		auth.POST("/user/delete-by-username", userCtrl.DeleteByUsername)  // 改为POST
	}
}
