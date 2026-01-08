package user

import (
	"devops/middleware"

	"github.com/gin-gonic/gin"
)

// SetupLogRoutes 设置日志管理路由
func SetupLogRoutes(r *gin.RouterGroup) {
	// TODO: 创建日志控制器后取消注释
	// operLogCtrl := userctrl.NewOperationLogController()
	// loginLogCtrl := userctrl.NewLoginLogController()

	// 需要JWT认证
	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	{
		// 操作日志
		// operLogs := auth.Group("/operation-logs")
		// {
		// 	operLogs.GET("", operLogCtrl.GetList)
		// 	operLogs.DELETE("/:id", operLogCtrl.Delete)
		// 	operLogs.DELETE("/clear", operLogCtrl.Clear)
		// }

		// 登录日志
		// loginLogs := auth.Group("/login-logs")
		// {
		// 	loginLogs.GET("", loginLogCtrl.GetList)
		// 	loginLogs.DELETE("/:id", loginLogCtrl.Delete)
		// 	loginLogs.DELETE("/clear", loginLogCtrl.Clear)
		// }
	}
}
