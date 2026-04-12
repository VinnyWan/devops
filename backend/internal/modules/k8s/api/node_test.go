package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"devops-platform/internal/modules/k8s/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type swaggerParameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}

type swaggerSchema struct {
	Ref        string                   `json:"$ref"`
	AllOf      []swaggerSchema          `json:"allOf"`
	Properties map[string]swaggerSchema `json:"properties"`
}

type swaggerResponse struct {
	Schema swaggerSchema `json:"schema"`
}

type swaggerOperation struct {
	Parameters []swaggerParameter          `json:"parameters"`
	Responses  map[string]swaggerResponse  `json:"responses"`
}

type swaggerPathItem struct {
	Get swaggerOperation `json:"get"`
}

type swaggerDefinition struct {
	Properties map[string]swaggerSchema `json:"properties"`
}

type swaggerDoc struct {
	Paths       map[string]swaggerPathItem    `json:"paths"`
	Definitions map[string]swaggerDefinition  `json:"definitions"`
}

func openNodeAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skipf("skip db-backed API test: %v", err)
		}
		t.Fatalf("open db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Cluster{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

func loadNodeSwaggerDoc(t *testing.T) swaggerDoc {
	t.Helper()
	content, err := os.ReadFile("../../../../docs/swagger/swagger.json")
	if err != nil {
		t.Fatalf("read swagger failed: %v", err)
	}

	var doc swaggerDoc
	if err := json.Unmarshal(content, &doc); err != nil {
		t.Fatalf("unmarshal swagger failed: %v", err)
	}
	return doc
}

func assertQueryParameter(t *testing.T, params []swaggerParameter, name string, required bool) {
	t.Helper()
	for _, param := range params {
		if param.Name == name && param.In == "query" {
			if param.Required != required {
				t.Fatalf("expected parameter %s required=%v, got %v", name, required, param.Required)
			}
			return
		}
	}
	t.Fatalf("expected query parameter %s", name)
}

func assertNodeDetail200Schema(t *testing.T, op swaggerOperation) {
	t.Helper()
	resp, ok := op.Responses["200"]
	if !ok {
		t.Fatalf("expected 200 response")
	}
	if len(resp.Schema.AllOf) == 0 {
		t.Fatalf("expected 200 schema allOf wrapper, got %#v", resp.Schema)
	}
	for _, item := range resp.Schema.AllOf {
		if data, ok := item.Properties["data"]; ok {
			if data.Ref != "#/definitions/service.NodeDetail" {
				t.Fatalf("expected data ref service.NodeDetail, got %s", data.Ref)
			}
			return
		}
	}
	t.Fatalf("expected 200 schema to include data property ref to service.NodeDetail")
}

func TestGetNodeDetailRequiresName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/detail?clusterName=api-t1-default", nil)

	GetNodeDetail(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name 不能为空") {
		t.Fatalf("expected missing name message, got %s", w.Body.String())
	}
}

func TestGetNodeDetailRejectsCrossTenantCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openNodeAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(11)
	tenant2 := uint(22)
	seedK8sTenantClusters(t, db, tenant1, tenant2)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/detail?clusterName=api-t2-default&name=node-a", nil)
	c.Set("tenantID", tenant1)

	GetNodeDetail(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "不存在或无权限") {
		t.Fatalf("expected tenant isolation error, got %s", w.Body.String())
	}
}

func TestGetNodeDetailUsesDefaultClusterFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openNodeAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(11)
	tenant2 := uint(22)
	clusters := seedK8sTenantClusters(t, db, tenant1, tenant2)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/detail?name=node-a", nil)
	c.Set("tenantID", tenant1)

	GetNodeDetail(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 from uninitialized client after fallback, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "服务初始化失败") {
		t.Fatalf("expected service init failure after fallback, got %s", w.Body.String())
	}

	resolved, err := resolveClusterName(c)
	if err != nil {
		t.Fatalf("expected fallback cluster resolved, got err=%v", err)
	}
	if resolved != clusters[0].Name {
		t.Fatalf("expected fallback cluster %s, got %s", clusters[0].Name, resolved)
	}
}

func TestGetNodeEventsRequiresName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/events?clusterName=api-t1-default", nil)

	GetNodeEvents(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name 不能为空") {
		t.Fatalf("expected missing name message, got %s", w.Body.String())
	}
}

func TestGetNodeEventsRejectsCrossTenantCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openNodeAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(11)
	tenant2 := uint(22)
	seedK8sTenantClusters(t, db, tenant1, tenant2)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/events?clusterName=api-t2-default&name=node-a", nil)
	c.Set("tenantID", tenant1)

	GetNodeEvents(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "不存在或无权限") {
		t.Fatalf("expected tenant isolation error, got %s", w.Body.String())
	}
}

func TestGetNodeEventsUsesDefaultClusterFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openNodeAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(11)
	tenant2 := uint(22)
	clusters := seedK8sTenantClusters(t, db, tenant1, tenant2)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/node/events?name=node-a", nil)
	c.Set("tenantID", tenant1)

	GetNodeEvents(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 from uninitialized client after fallback, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "服务初始化失败") {
		t.Fatalf("expected service init failure after fallback, got %s", w.Body.String())
	}

	resolved, err := resolveClusterName(c)
	if err != nil {
		t.Fatalf("expected fallback cluster resolved, got err=%v", err)
	}
	if resolved != clusters[0].Name {
		t.Fatalf("expected fallback cluster %s, got %s", clusters[0].Name, resolved)
	}
}

func TestNodeSwaggerContract(t *testing.T) {
	doc := loadNodeSwaggerDoc(t)

	detailPath, ok := doc.Paths["/k8s/node/detail"]
	if !ok {
		t.Fatalf("expected swagger path /k8s/node/detail")
	}
	assertQueryParameter(t, detailPath.Get.Parameters, "clusterName", false)
	assertQueryParameter(t, detailPath.Get.Parameters, "name", true)
	assertNodeDetail200Schema(t, detailPath.Get)

	detailDef, ok := doc.Definitions["service.NodeDetail"]
	if !ok {
		t.Fatalf("expected definition service.NodeDetail")
	}
	for _, key := range []string{"annotations", "lease", "capacity", "allocatable", "podCIDR", "providerID", "pods", "allocatedResources"} {
		if _, ok := detailDef.Properties[key]; !ok {
			t.Fatalf("expected service.NodeDetail property %q", key)
		}
	}

	eventsPath, ok := doc.Paths["/k8s/node/events"]
	if !ok {
		t.Fatalf("expected swagger path /k8s/node/events")
	}
	assertQueryParameter(t, eventsPath.Get.Parameters, "clusterName", false)
	assertQueryParameter(t, eventsPath.Get.Parameters, "name", true)
}
