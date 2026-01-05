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
	"devops/internal/logger"
	"devops/internal/middleware"
	"devops/routers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "c", "", "配置文件路径 (默认: ./config.yaml)")
	flag.Parse()

	// 加载配置文件
	if err := config.LoadConfig(configPath); err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 初始化 zap（JSON 格式）
	logger.Init()
	defer logger.Log.Sync() // 程序退出前刷新缓冲

	// 设置 Gin 运行模式
	gin.SetMode(config.Config.Server.Model)

	// 从 routers 包加载路由（返回 *gin.Engine）
	router := routers.SetupRouter()

	// 使用 zap 中间件，接管所有请求日志和 panic
	router.Use(
		middleware.GinZapLogger(logger.Log),
		middleware.GinRecoveryWithZap(logger.Log),
	)

	// 创建 HTTP 服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Config.Server.Port),
		Handler: router,
	}

	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("服务启动失败", zap.Error(err))
		}
	}()
	logger.Log.Info("Gin 服务已启动",
		zap.String("port", config.Config.Server.Port),
	)

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("服务器关闭失败", zap.Error(err))
	} else {
		logger.Log.Info("服务器已优雅停止")
	}
}
