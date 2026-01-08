package routers

import (
	_ "devops/docs"
	"devops/internal/logger"
	"devops/middleware"
	k8srouters "devops/routers/k8s"
	userrouters "devops/routers/user"
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
		// 用户模块路由
		userrouters.SetupAuthRoutes(api)
		userrouters.SetupUserRoutes(api)
		userrouters.SetupRoleRoutes(api)
		userrouters.SetupMenuRoutes(api)
		userrouters.SetupDepartmentRoutes(api)
		userrouters.SetupPostRoutes(api)
		userrouters.SetupLogRoutes(api)

		// K8s模块路由
		k8srouters.SetupClusterRoutes(api)
		k8srouters.SetupResourceRoutes(api)
	}

	return r
}
