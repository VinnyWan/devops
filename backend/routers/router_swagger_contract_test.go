package routers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"devops-platform/config"
	"devops-platform/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type swaggerDoc struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type openapiDoc struct {
	OpenAPI string                                `json:"openapi"`
	Paths   map[string]map[string]json.RawMessage `json:"paths"`
}

func setupRouterForSwaggerTest() *gin.Engine {
	v := viper.New()
	v.Set("server.mode", "test")
	v.Set("cors.allow_origins", []string{"http://localhost:3000"})
	config.Cfg = v
	logger.Log = zap.NewNop()
	return InitRouter()
}

func TestSwaggerUIEndpoints(t *testing.T) {
	router := setupRouterForSwaggerTest()

	indexReq := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	indexRes := httptest.NewRecorder()
	router.ServeHTTP(indexRes, indexReq)
	if indexRes.Code != http.StatusOK {
		t.Fatalf("swagger index expected %d, got %d", http.StatusOK, indexRes.Code)
	}

	docReq := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	docRes := httptest.NewRecorder()
	router.ServeHTTP(docRes, docReq)
	if docRes.Code != http.StatusOK {
		t.Fatalf("swagger doc expected %d, got %d", http.StatusOK, docRes.Code)
	}

	var doc swaggerDoc
	if err := json.Unmarshal(docRes.Body.Bytes(), &doc); err != nil {
		t.Fatalf("swagger doc json unmarshal failed: %v", err)
	}
	if len(doc.Paths) == 0 {
		t.Fatalf("swagger doc paths should not be empty")
	}
}

func TestSwaggerPathsCoveredByRouter(t *testing.T) {
	router := setupRouterForSwaggerTest()

	docReq := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	docRes := httptest.NewRecorder()
	router.ServeHTTP(docRes, docReq)
	if docRes.Code != http.StatusOK {
		t.Fatalf("swagger doc expected %d, got %d", http.StatusOK, docRes.Code)
	}

	var doc swaggerDoc
	if err := json.Unmarshal(docRes.Body.Bytes(), &doc); err != nil {
		t.Fatalf("swagger doc json unmarshal failed: %v", err)
	}

	registered := make(map[string]map[string]bool)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/") {
			continue
		}
		if strings.EqualFold(route.Method, http.MethodOptions) || strings.EqualFold(route.Method, http.MethodHead) {
			continue
		}
		docPath := strings.TrimPrefix(route.Path, "/api/v1")
		docPath = normalizeGinPathToSwaggerPath(docPath)
		methodBucket := registered[docPath]
		if methodBucket == nil {
			methodBucket = make(map[string]bool)
			registered[docPath] = methodBucket
		}
		methodBucket[strings.ToLower(route.Method)] = true
	}

	var mismatches []string
	for docPath, methodMap := range doc.Paths {
		routeMethods, exists := registered[docPath]
		if !exists {
			mismatches = append(mismatches, "missing route path "+docPath)
			continue
		}
		if len(methodMap) == 0 || len(routeMethods) == 0 {
			mismatches = append(mismatches, "missing route or swagger methods "+docPath)
		}
	}

	if len(mismatches) > 0 {
		t.Fatalf("swagger/router mismatch count=%d, first=%s", len(mismatches), mismatches[0])
	}
}

func TestOpenAPIFileVersionAndPaths(t *testing.T) {
	content, err := os.ReadFile("../docs/openapi/openapi.json")
	if err != nil {
		t.Fatalf("read openapi file failed: %v", err)
	}

	var doc openapiDoc
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("openapi file json unmarshal failed: %v", err)
	}
	if !strings.HasPrefix(doc.OpenAPI, "3.") {
		t.Fatalf("openapi version expected 3.x, got %q", doc.OpenAPI)
	}
	if len(doc.Paths) == 0 {
		t.Fatalf("openapi doc paths should not be empty")
	}
}

func TestOpenAPIPathsCoveredByRouter(t *testing.T) {
	router := setupRouterForSwaggerTest()

	content, err := os.ReadFile("../docs/openapi/openapi.json")
	if err != nil {
		t.Fatalf("read openapi file failed: %v", err)
	}

	var doc openapiDoc
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("openapi file json unmarshal failed: %v", err)
	}

	registered := make(map[string]map[string]bool)
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/") {
			continue
		}
		if strings.EqualFold(route.Method, http.MethodOptions) || strings.EqualFold(route.Method, http.MethodHead) {
			continue
		}
		docPath := strings.TrimPrefix(route.Path, "/api/v1")
		docPath = normalizeGinPathToSwaggerPath(docPath)
		methodBucket := registered[docPath]
		if methodBucket == nil {
			methodBucket = make(map[string]bool)
			registered[docPath] = methodBucket
		}
		methodBucket[strings.ToLower(route.Method)] = true
	}

	var mismatches []string
	for docPath, methodMap := range doc.Paths {
		routeMethods, exists := registered[docPath]
		if !exists {
			mismatches = append(mismatches, "missing route path "+docPath)
			continue
		}
		if len(methodMap) == 0 || len(routeMethods) == 0 {
			mismatches = append(mismatches, "missing route or openapi methods "+docPath)
		}
	}

	if len(mismatches) > 0 {
		t.Fatalf("openapi/router mismatch count=%d, first=%s", len(mismatches), mismatches[0])
	}
}

var ginParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

func normalizeGinPathToSwaggerPath(pathValue string) string {
	return ginParamPattern.ReplaceAllString(pathValue, `{$1}`)
}
