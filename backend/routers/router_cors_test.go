package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-platform/config"

	"github.com/spf13/viper"
)

func TestCORSPreflight(t *testing.T) {
	v := viper.New()
	v.Set("cors.allow_origins", []string{"http://localhost:3000"})
	config.Cfg = v

	r := InitRouter()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/k8s/cluster/nodes?id=6&page=1&pageSize=10", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow-origin %q, got %q", "http://localhost:3000", got)
	}

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected allow-credentials true, got %q", got)
	}
}
