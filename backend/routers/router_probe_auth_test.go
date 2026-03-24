package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-platform/config"
	"devops-platform/internal/pkg/logger"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestHealthAndReadyEndpoints(t *testing.T) {
	v := viper.New()
	v.Set("server.mode", "test")
	v.Set("cors.allow_origins", []string{"http://localhost:3000"})
	config.Cfg = v
	logger.Log = zap.NewNop()

	r := InitRouter()

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthW := httptest.NewRecorder()
	r.ServeHTTP(healthW, healthReq)
	if healthW.Code != http.StatusOK {
		t.Fatalf("healthz expected %d, got %d", http.StatusOK, healthW.Code)
	}

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyW := httptest.NewRecorder()
	r.ServeHTTP(readyW, readyReq)
	if readyW.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz expected %d, got %d", http.StatusServiceUnavailable, readyW.Code)
	}
}

func TestUnauthorizedK8sRoute(t *testing.T) {
	v := viper.New()
	v.Set("server.mode", "test")
	v.Set("cors.allow_origins", []string{"http://localhost:3000"})
	config.Cfg = v
	logger.Log = zap.NewNop()

	r := InitRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/k8s/cluster/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestOIDCCallbackMissingState(t *testing.T) {
	v := viper.New()
	v.Set("server.mode", "test")
	v.Set("cors.allow_origins", []string{"http://localhost:3000"})
	config.Cfg = v
	logger.Log = zap.NewNop()

	r := InitRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}
