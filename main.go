package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devops/common/config"
	"devops/internal/database"
	"devops/internal/logger"
	"devops/routers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title DevOps 系统管理平台 API
// @version 1.0
// @description 这是一个基于Gin框架的DevOps系统管理平台API文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入Bearer Token，格式：Bearer {token}，注意中间有空格

func main() {
	// 1. 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "c", "", "配置文件路径 (默认: ./config.yaml)")
	flag.Parse()

	// 2. 加载配置文件
	resolvedConfigPath := configPath
	if resolvedConfigPath == "" {
		resolvedConfigPath = "./config.yaml"
	}
	if err := config.LoadConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 初始化日志
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Log.Sync()
	}()

	// 记录配置信息
	logger.Log.Info("配置文件加载成功",
		zap.String("config_path", resolvedConfigPath),
		zap.String("server_port", config.Config.Server.Port),
		zap.String("server_mode", config.Config.Server.Model),
		zap.Bool("swagger_enabled", config.Config.Server.EnableSwagger),
		zap.String("db_host", config.Config.Db.Host),
		zap.Int("db_port", config.Config.Db.Port),
		zap.String("db_name", config.Config.Db.Db),
		zap.String("redis_address", config.Config.Redis.Address),
	)

	// 4. 初始化DB
	if err := database.InitMysql(); err != nil {
		logger.Log.Fatal("Mysql数据库初始化失败", zap.Error(err))
	}

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		logger.Log.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 初始化基础数据
	if err := database.InitData(); err != nil {
		logger.Log.Fatal("初始化数据失败", zap.Error(err))
	}

	// 5. 初始化Redis
	if err := database.InitRedis(); err != nil {
		logger.Log.Fatal("Redis数据库初始化失败", zap.Error(err))
	}
	// 5. 初始化 Gin
	gin.SetMode(config.Config.Server.Model)
	router := setupRouter()

	// 6. 创建并启动 HTTP 服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Config.Server.Port),
		Handler: router,
	}

	startHTTPServer(srv)

	// 7. 优雅关闭
	gracefulShutdown(srv)
}

// setupRouter 初始化路由
func setupRouter() *gin.Engine {
	// 路由已在 routers.SetupRouter() 中应用了全局中间件
	return routers.SetupRouter()
}

// startHTTPServer 启动 HTTP 服务
func startHTTPServer(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("服务启动失败", zap.Error(err))
		}
	}()
	logger.Log.Info("Gin 服务已启动",
		zap.String("addr", srv.Addr),
	)
}

// gracefulShutdown 优雅退出
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // 阻塞，直到收到信号

	logger.Log.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("服务器关闭失败", zap.Error(err))
	} else {
		logger.Log.Info("服务器已优雅停止")
	}
}
