package database

import (
	"context"
	"devops/common/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client // 全局 Redis 客户端
)

// Init 初始化 Redis 客户端（在 main 中调用一次）
func InitRedis() error {
	rc := config.GetRedisConfig() // *config.redis
	if rc.Address == "" {
		return fmt.Errorf("redis address is empty")
	}

	poolSize := rc.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}
	minIdle := rc.MinIdleConns
	if minIdle < 0 {
		minIdle = 0
	}
	if minIdle == 0 && poolSize >= 5 {
		minIdle = 2
	}

	// 创建客户端
	Client = redis.NewClient(&redis.Options{
		Addr:         rc.Address,
		Password:     rc.Password,
		DB:           rc.DB,    // 默认库，可按需配置
		PoolSize:     poolSize, // 连接池大小，可根据需求调整
		MinIdleConns: minIdle,  // 最小空闲连接数
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
