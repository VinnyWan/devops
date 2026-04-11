package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-platform/config"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/service"
	"devops-platform/internal/pkg/logger"
	redisPkg "devops-platform/internal/pkg/redis"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSessionAuthRejectsSessionWithoutTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupSessionTestConfig()
	setupSessionTestRedis(t)

	testDB, err := gorm.Open(sqlite.Open("file:session_auth_missing_tenant?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := testDB.AutoMigrate(&model.Tenant{}); err != nil {
		t.Fatalf("migrate tenant failed: %v", err)
	}
	db = testDB
	t.Cleanup(func() {
		db = nil
	})

	sessionID, err := service.NewSessionService().CreateSession(context.Background(), 1, "tester", 0, "", "local", "", "")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/info", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	c.Request = req

	SessionAuth()(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

func TestSessionAuthFailsWhenDBNotInitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupSessionTestConfig()
	setupSessionTestRedis(t)
	db = nil

	sessionID, err := service.NewSessionService().CreateSession(context.Background(), 3, "tester3", 1, "tenant-a", "local", "", "")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/info", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	c.Request = req

	SessionAuth()(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
}

func TestSessionAuthRejectsInactiveTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupSessionTestConfig()
	setupSessionTestRedis(t)

	testDB, err := gorm.Open(sqlite.Open("file:session_auth_inactive_tenant?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := testDB.AutoMigrate(&model.Tenant{}); err != nil {
		t.Fatalf("migrate tenant failed: %v", err)
	}
	tenant := model.Tenant{Name: "inactive-tenant", Code: "inactive-tenant", Status: "inactive"}
	if err := testDB.Create(&tenant).Error; err != nil {
		t.Fatalf("create tenant failed: %v", err)
	}
	db = testDB
	t.Cleanup(func() {
		db = nil
	})

	sessionID, err := service.NewSessionService().CreateSession(context.Background(), 2, "tester2", tenant.ID, tenant.Code, "local", "", "")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/info", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	c.Request = req

	SessionAuth()(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

func setupSessionTestConfig() {
	v := viper.New()
	v.Set("server.mode", "test")
	config.Cfg = v
	logger.Log = zap.NewNop()
}

func setupSessionTestRedis(t *testing.T) {
	t.Helper()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	redisPkg.Client = redisv9.NewClient(&redisv9.Options{Addr: mini.Addr()})

	t.Cleanup(func() {
		_ = redisPkg.Client.Close()
		mini.Close()
	})
}
