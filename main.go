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
	"devops/middleware"
	"devops/routers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "c", "", "配置文件路径 (默认: ./config.yaml)")
	flag.Parse()

	// 2. 加载配置文件
	// 加载配置文件
	if err := config.LoadConfig(configPath); err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 3. 初始化日志
	logger.Init()
	defer logger.Log.Sync()

	// 记录配置信息
	logger.Log.Info("配置文件加载成功",
		zap.String("config_path", configPath),
		zap.Any("server", config.Config.Server),
		zap.Any("db", config.Config.Db),
		zap.Any("redis", config.Config.Redis),
	)

	// 4. 初始化DB
	if err := database.InitMysql(); err != nil {
		logger.Log.Fatal("Mysql数据库初始化失败", zap.Error(err))
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

// setupRouter 初始化路由和中间件
func setupRouter() *gin.Engine {
	router := routers.SetupRouter()

	// 全局中间件：请求日志 + panic 恢复
	router.Use(
		middleware.GinZapLogger(logger.Log),
		middleware.GinRecoveryWithZap(logger.Log),
	)

	return router
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
