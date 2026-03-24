package bootstrap

import (
	"devops-platform/config"
	"devops-platform/internal/pkg/redis"
	"fmt"
)

func InitRedis() error {
	addr := config.Cfg.GetString("redis.addr")
	password := config.Cfg.GetString("redis.password")
	db := config.Cfg.GetInt("redis.db")

	if addr == "" {
		// 尝试默认值，防止 :0 错误
		addr = "127.0.0.1:6379"
		fmt.Println("Warning: redis addr is empty, using default 127.0.0.1:6379")
	}

	fmt.Printf("Connecting to Redis: %s (db: %d)\n", addr, db)

	return redis.Init(addr, password, db)
}
