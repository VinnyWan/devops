package routers

import (
	"devops/controller"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes 设置认证相关路由
func SetupAuthRoutes(r *gin.RouterGroup) {
	userCtrl := controller.NewUserController()
	captchaCtrl := controller.NewCaptchaController()

	// 验证码
	r.GET("/captcha", captchaCtrl.Generate)
	r.GET("/captcha/:id", captchaCtrl.Serve)

	// 认证
	auth := r.Group("/auth")
	{
		auth.POST("/login", userCtrl.Login)
	}
}
