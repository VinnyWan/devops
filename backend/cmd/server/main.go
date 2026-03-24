package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devops-platform/config"
	"devops-platform/internal/bootstrap"
	"devops-platform/internal/middleware"
	k8sAPI "devops-platform/internal/modules/k8s/api"
	userAPI "devops-platform/internal/modules/user/api"
	"devops-platform/internal/modules/user/repository"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"
	"devops-platform/internal/pkg/utils"
	"devops-platform/routers"

	"go.uber.org/zap"
)

// @title DevOps 运维平台 API
// @version 1.0
// @description DevOps 运维平台接口文档，登录后在 Authorize 中输入 Bearer {session_id} 即可调用接口
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {session_id}，session_id 从登录接口获取 (Swagger 暂不支持 Cookie 自动发送，请手动输入)
func main() {
	// 阶段1：使用默认配置启动临时 Logger（配置加载前）
	if err := logger.Init(); err != nil {
		log.Fatal("Logger init failed:", err)
	}
	defer logger.Log.Sync()

	// 阶段2：加载配置
	if err := bootstrap.InitConfig(); err != nil {
		log.Fatal(err)
	}

	// 阶段3：用配置重新初始化 Logger
	if err := logger.InitWithConfig(&logger.LogConfig{
		Level:            config.Cfg.GetString("log.level"),
		Output:           config.Cfg.GetString("log.output"),
		FilePath:         config.Cfg.GetString("log.filePath"),
		EnableCaller:     config.Cfg.GetBool("log.enableCaller"),
		EnableStacktrace: config.Cfg.GetBool("log.enableStacktrace"),
	}); err != nil {
		log.Fatal("Logger reinit failed:", err)
	}

	if err := utils.InitCrypto(); err != nil {
		log.Fatal("加密模块初始化失败:", err)
	}
	if err := bootstrap.InitDB(); err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.InitRedis(); err != nil {
		log.Fatal("Redis init failed:", err)
	}
	// 设置用户 API 的 DB 实例
	userAPI.SetDB(bootstrap.DB)
	// 设置中间件的 DB 实例
	middleware.SetDB(bootstrap.DB)

	// 开启审计日志清理任务
	auditRepo := repository.NewAuditRepo(bootstrap.DB)
	auditService := service.NewAuditService(auditRepo)
	auditService.StartAuditCleanupTask()
	if err := bootstrap.InitCasbin(); err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.InitK8sFactory(); err != nil {
		log.Fatal(err)
	}
	// 设置 K8s 服务的 DB 实例和客户端工厂
	k8sAPI.SetK8sDB(bootstrap.DB, bootstrap.K8sFactory)

	r := routers.InitRouter()
	port := config.Cfg.GetString("server.port")
	if port == "" {
		port = "8000"
	}
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	registration, err := config.RegisterToNacos(config.Cfg)
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info(fmt.Sprintf("服务启动，监听地址: %s", addr))
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if registration != nil {
		if err := registration.Deregister(); err != nil {
			logger.Log.Warn("Nacos 实例注销失败", zap.Error(err))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
