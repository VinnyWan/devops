package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(addr, password string, db int) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	return err
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Del(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}

func HSet(ctx context.Context, key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return Client.SAdd(ctx, key, members...).Err()
}

func SRem(ctx context.Context, key string, members ...interface{}) error {
	return Client.SRem(ctx, key, members...).Err()
}

func SMembers(ctx context.Context, key string) ([]string, error) {
	return Client.SMembers(ctx, key).Result()
}

func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}
