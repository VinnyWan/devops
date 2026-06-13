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
	jwtpkg "devops-platform/internal/pkg/jwt"
	"devops-platform/internal/pkg/logger"
	"devops-platform/internal/pkg/utils"
	"devops-platform/routers"

	"go.uber.org/zap"
)

// @title DevOps 运维平台 API
// @version 1.0
// @description DevOps 运维平台接口文档，支持 Session Cookie / JWT Bearer / API Key 三种认证方式
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
// @description 支持三种格式: Bearer <jwt_token> | Bearer <session_id> | ApiKey <api_key>
func main() {
	// Phase 1: Bootstrap logger with defaults (before config)
	if err := logger.Init(); err != nil {
		log.Fatal("Logger init failed:", err)
	}
	defer logger.Log.Sync()

	// Phase 2: Load configuration
	if err := bootstrap.InitConfig(); err != nil {
		log.Fatal(err)
	}

	// Phase 3: Reinit logger with config values
	if err := logger.InitWithConfig(&logger.LogConfig{
		Level:            config.Cfg.GetString("log.level"),
		Output:           config.Cfg.GetString("log.output"),
		FilePath:         config.Cfg.GetString("log.filePath"),
		EnableCaller:     config.Cfg.GetBool("log.enableCaller"),
		EnableStacktrace: config.Cfg.GetBool("log.enableStacktrace"),
	}); err != nil {
		log.Fatal("Logger reinit failed:", err)
	}

	// Phase 4: Initialize crypto and JWT
	if err := utils.InitCrypto(); err != nil {
		log.Fatal("Crypto module init failed:", err)
	}
	jwtSecret := config.Cfg.GetString("crypto.secret")
	if jwtSecret == "" {
		jwtSecret = "devops-platform-default-jwt-secret"
	}
	jwtpkg.InitDefault(jwtSecret)
	logger.Log.Info("JWT module initialized")

	// Phase 5: Initialize infrastructure (DB, Redis, Casbin, K8s)
	if err := bootstrap.InitDB(); err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.InitRedis(); err != nil {
		log.Fatal("Redis init failed:", err)
	}
	if err := bootstrap.InitCasbin(bootstrap.DB); err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.InitK8sFactory(); err != nil {
		log.Fatal(err)
	}

	// Phase 6: Initialize all business modules (single call, all wiring centralized)
	bootstrap.InitModules(bootstrap.DB)
	bootstrap.StartModuleBackgroundTasks()

	// Phase 7: Start HTTP server
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
	logger.Log.Info(fmt.Sprintf("Server listening on %s", addr))
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Phase 8: Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if registration != nil {
		if err := registration.Deregister(); err != nil {
			logger.Log.Warn("Nacos deregister failed", zap.Error(err))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
