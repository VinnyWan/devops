package user

import (
	"context"
	"fmt"
	"time"

	"devops/internal/database"

	"github.com/dchest/captcha"
)

// CaptchaService 验证码服务
type CaptchaService struct{}

// RedisStore 基于Redis的验证码存储
type RedisStore struct{}

// Set 存储验证码
func (s RedisStore) Set(id string, digits []byte) {
	ctx := context.Background()
	key := fmt.Sprintf("captcha:%s", id)
	database.Client.Set(ctx, key, string(digits), 5*time.Minute)
}

// Get 获取验证码
func (s RedisStore) Get(id string, clear bool) []byte {
	ctx := context.Background()
	key := fmt.Sprintf("captcha:%s", id)

	val, err := database.Client.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	if clear {
		database.Client.Del(ctx, key)
	}

	return []byte(val)
}

var store = RedisStore{}

func init() {
	// 设置自定义存储
	captcha.SetCustomStore(store)
}

// Generate 生成验证码
func (s *CaptchaService) Generate() (id string, err error) {
	// 生成验证码（会自动存储到RedisStore）
	id = captcha.New()
	return id, nil
}

// Verify 验证验证码
func (s *CaptchaService) Verify(id, code string) bool {
	// 验证并清除验证码
	return captcha.VerifyString(id, code)
}
