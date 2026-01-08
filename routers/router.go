package routers

import (
	_ "devops/docs"
	"devops/internal/logger"
	"devops/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置总路由，汇总所有模块路由
func SetupRouter() *gin.Engine {
	r := gin.New() // 不用 gin.Default()

	// CORS中间件配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有来源，生产环境应改为具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 全局中间件：日志记录 + Panic恢复
	r.Use(
		middleware.GinZapLogger(logger.Log),
		middleware.GinRecoveryWithZap(logger.Log),
	)

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由（登录、验证码）
		SetupAuthRoutes(api)

		// 用户管理路由
		SetupUserRoutes(api)

		// 角色管理路由
		SetupRoleRoutes(api)

		// 菜单管理路由
		SetupMenuRoutes(api)

		// 部门管理路由
		SetupDepartmentRoutes(api)

		// 岗位管理路由
		SetupPostRoutes(api)

		// 日志管理路由
		SetupLogRoutes(api)
	}

	return r
}
