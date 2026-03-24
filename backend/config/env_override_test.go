package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("DEVOPS_DB_PASSWORD", "from-env")
	t.Setenv("DEVOPS_REDIS_PASSWORD", "redis-env")
	t.Setenv("DEVOPS_JWT_SECRET", "jwt-env")

	v := viper.New()
	LoadEnvOverrides(v)

	if got := v.GetString("db.password"); got != "from-env" {
		t.Fatalf("expected db.password from env, got %q", got)
	}
	if got := v.GetString("redis.password"); got != "redis-env" {
		t.Fatalf("expected redis.password from env, got %q", got)
	}
	if got := v.GetString("jwt.secret"); got != "jwt-env" {
		t.Fatalf("expected jwt.secret from env, got %q", got)
	}
}
