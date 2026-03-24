package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"devops-platform/internal/modules/user/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuditTestDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := testDB.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db = testDB
	auditChannel = make(chan *model.AuditLog, 128)
	go auditWorker()
	return testDB
}

func waitAuditLog(t *testing.T, testDB *gorm.DB, path string) model.AuditLog {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var log model.AuditLog
		err := testDB.Where("path = ?", path).Order("id DESC").First(&log).Error
		if err == nil {
			return log
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("audit log not found for path: %s", path)
	return model.AuditLog{}
}

func TestAuditRecordsGetRequestWithMaskedQuery(t *testing.T) {
	testDB := setupAuditTestDB(t, "file:audit_middleware_get?mode=memory&cache=shared")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext(), Audit())
	r.GET("/api/v1/test", func(c *gin.Context) {
		c.Set("userID", uint(1001))
		c.Set("username", "tester-get")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test?password=123456&keyword=abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	log := waitAuditLog(t, testDB, "/api/v1/test")
	if log.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %s", log.Method)
	}
	if !strings.Contains(log.Params, `"password":"***"`) {
		t.Fatalf("expected masked query password, got %s", log.Params)
	}
	if !strings.Contains(log.Params, `"keyword":"abc"`) {
		t.Fatalf("expected query keyword recorded, got %s", log.Params)
	}
	if log.RequestAt.IsZero() {
		t.Fatalf("expected requestAt recorded")
	}
}

func TestAuditRecordsPostRequestAndResponseWithMasking(t *testing.T) {
	testDB := setupAuditTestDB(t, "file:audit_middleware_post?mode=memory&cache=shared")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContext(), Audit())
	r.POST("/api/v1/login", func(c *gin.Context) {
		c.Set("userID", uint(1002))
		c.Set("username", "tester-post")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
			"token": "server-token",
		})
	})

	body := `{"username":"u1","password":"abc","token":"client-token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	log := waitAuditLog(t, testDB, "/api/v1/login")
	if log.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", log.Method)
	}
	if !strings.Contains(log.Params, `"password":"***"`) {
		t.Fatalf("expected masked request password, got %s", log.Params)
	}
	if !strings.Contains(log.Params, `"token":"***"`) {
		t.Fatalf("expected masked request token, got %s", log.Params)
	}
	if !strings.Contains(log.Result, `"status":401`) {
		t.Fatalf("expected response status in result, got %s", log.Result)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(log.Result), &resultMap); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	bodyValue, ok := resultMap["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected response body in result, got %#v", resultMap["body"])
	}
	if token, ok := bodyValue["token"].(string); !ok || token != "***" {
		t.Fatalf("expected masked response token, got %#v", bodyValue["token"])
	}
}
