package user

import (
	userctrl "devops/controller/user"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes 设置认证相关路由
func SetupAuthRoutes(r *gin.RouterGroup) {
	userCtrl := userctrl.NewUserController()

	// 认证
	auth := r.Group("/auth")
	{
		auth.POST("/login", userCtrl.Login)
	}
}
