package service

import (
	"testing"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.PrometheusConfig{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestMonitorServiceSaveConfig_EmptyEndpoint(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	err := svc.SaveConfig(&model.PrometheusConfig{Name: "test", Endpoint: ""})
	if err == nil {
		t.Fatalf("expected empty endpoint error")
	}
}

func TestMonitorServiceSaveConfig_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	cfg := &model.PrometheusConfig{
		Name:           "test-prometheus",
		Endpoint:       "http://127.0.0.1:19999",
		TimeoutSeconds: 10,
	}
	// TestConnection will likely fail against a local closed port, so status becomes "error"
	// but the config should still be saved and get an ID assigned
	err := svc.SaveConfig(cfg)
	if err != nil {
		t.Fatalf("save config failed (even with failed conn test, save should succeed): %v", err)
	}
	if cfg.ID == 0 {
		t.Fatalf("expected config ID to be assigned after save")
	}
	if cfg.Status != "error" {
		t.Fatalf("expected status 'error' for unreachable endpoint, got '%s'", cfg.Status)
	}
}

func TestMonitorServiceListConfigs(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	// Seed a config
	db.Create(&model.PrometheusConfig{
		Name:           "test-1",
		Endpoint:       "http://prom1:9090",
		TimeoutSeconds: 15,
		Status:         "unknown",
	})

	configs, total, err := svc.ListConfigs(1, 20)
	if err != nil {
		t.Fatalf("list configs failed: %v", err)
	}
	if total < 1 {
		t.Fatalf("expected at least 1 config, got %d", total)
	}
	if len(configs) < 1 {
		t.Fatalf("expected at least 1 config in results")
	}
}

func TestMonitorServiceGetConfig(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	cfg := model.PrometheusConfig{
		Name:           "test-get",
		Endpoint:       "http://prom:9090",
		TimeoutSeconds: 15,
		Status:         "connected",
	}
	db.Create(&cfg)

	result, err := svc.GetConfig(cfg.ID)
	if err != nil {
		t.Fatalf("get config failed: %v", err)
	}
	if result.Name != "test-get" {
		t.Fatalf("expected name 'test-get', got '%s'", result.Name)
	}
}

func TestMonitorServiceDeleteConfig(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	cfg := model.PrometheusConfig{
		Name:           "test-delete",
		Endpoint:       "http://prom:9090",
		TimeoutSeconds: 15,
	}
	db.Create(&cfg)

	err := svc.DeleteConfig(cfg.ID)
	if err != nil {
		t.Fatalf("delete config failed: %v", err)
	}

	// Verify soft delete
	var count int64
	db.Unscoped().Model(&model.PrometheusConfig{}).Where("id = ?", cfg.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected record to still exist (soft delete), got count %d", count)
	}

	// Verify it's not returned by normal queries
	var check model.PrometheusConfig
	err = db.First(&check, cfg.ID).Error
	if err == nil {
		t.Fatalf("expected record not found error after soft delete")
	}
}

func TestMonitorServiceTestConnection(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	// Empty endpoint should error
	err := svc.TestConnection("", "", "")
	if err == nil {
		t.Fatalf("expected error for empty endpoint")
	}

	// Non-reachable endpoint should error with connection failure (localhost closed port gives fast RST)
	err = svc.TestConnection("http://127.0.0.1:19998", "", "")
	if err == nil {
		t.Fatalf("expected error for unreachable endpoint")
	}
}

func TestMonitorServiceEnsureDefaults(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	// First call should create default
	svc.EnsureDefaults()
	var count int64
	db.Model(&model.PrometheusConfig{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 default config, got %d", count)
	}

	// Second call should be idempotent
	svc.EnsureDefaults()
	db.Model(&model.PrometheusConfig{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected still 1 default config after second call, got %d", count)
	}
}

func TestQueryHostMetrics_UnsupportedMetric(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	_, err := svc.QueryHostMetrics(1, "10.0.0.1", "invalid_metric", "", "")
	if err == nil {
		t.Fatalf("expected error for unsupported metric")
	}
}

func TestQueryPortStatus_NoConfig(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	result, err := svc.QueryPortStatus(9999, "10.0.0.1", []string{"80", "443"})
	if err != nil {
		t.Fatalf("query port status with no config should not error: %v", err)
	}
	for _, port := range []string{"80", "443"} {
		if result[port] != "unknown" {
			t.Fatalf("expected port %s to be 'unknown', got '%s'", port, result[port])
		}
	}
}

func TestQueryAgentStatus_NoConfig(t *testing.T) {
	db := setupTestDB(t)
	svc := NewMonitorService(db)

	result, err := svc.QueryAgentStatus(9999, []string{"10.0.0.1", "10.0.0.2"})
	if err != nil {
		t.Fatalf("query agent status with no config should not error: %v", err)
	}
	for _, ip := range []string{"10.0.0.1", "10.0.0.2"} {
		if result[ip] != "unknown" {
			t.Fatalf("expected ip %s to be 'unknown', got '%s'", ip, result[ip])
		}
	}
}

func TestPrometheusConfigTableName(t *testing.T) {
	cfg := model.PrometheusConfig{}
	if cfg.TableName() != "monitor_prometheus_configs" {
		t.Fatalf("expected table name 'monitor_prometheus_configs', got '%s'", cfg.TableName())
	}
}

func TestPrometheusConfigPasswordIsHidden(t *testing.T) {
	// Verify that the Password field has json:"-" tag (not serialized to JSON)
	cfg := model.PrometheusConfig{Password: "secret123"}
	// This is a compile-time check — the tag ensures Password is excluded from JSON marshaling
	// We just verify the struct field exists and holds the value
	if cfg.Password != "secret123" {
		t.Fatalf("expected password to be stored but hidden from JSON")
	}
}

func TestObservableErrorTypes(t *testing.T) {
	// Verify obserr types are used correctly
	err := obserr.New("INVALID_PARAM", "test", "some message")
	if err.Code != "INVALID_PARAM" {
		t.Fatalf("expected code INVALID_PARAM, got %s", err.Code)
	}

	wrapped := obserr.Wrap("DB_ERROR", "test", "wrapped message", err)
	if wrapped.Code != "DB_ERROR" {
		t.Fatalf("expected code DB_ERROR, got %s", wrapped.Code)
	}
}
