package routers

import (
	"context"
	"net/http"
	"time"

	"devops-platform/internal/pkg/redis"

	"github.com/gin-gonic/gin"

	"devops-platform/internal/middleware"
	v1 "devops-platform/routers/v1"

	_ "devops-platform/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(
		middleware.CORS(),
		middleware.Charset(), // 统一设置 UTF-8 字符集
		middleware.Recover(),
		middleware.RequestLog(),
	)

	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ready := true
		checks := gin.H{"app": "ok"}
		queued, dropped, capacity := middleware.AuditStats()
		checks["audit"] = gin.H{
			"queue": gin.H{
				"queued":   queued,
				"capacity": capacity,
			},
			"droppedTotal": dropped,
		}

		if redis.Client == nil {
			ready = false
			checks["redis"] = "not_initialized"
		} else {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := redis.Client.Ping(ctx).Err(); err != nil {
				ready = false
				checks["redis"] = "unhealthy"
			} else {
				checks["redis"] = "ok"
			}
		}

		if !ready {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"checks": checks,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": checks,
		})
	})

	// Swagger 路由 - 使用默认配置
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册 v1 路由
	v1.Register(r)

	return r
}
