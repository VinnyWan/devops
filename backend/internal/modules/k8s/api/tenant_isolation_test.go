package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devops-platform/internal/modules/k8s/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupK8sAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Cluster{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

func seedK8sTenantClusters(t *testing.T, db *gorm.DB, tenant1, tenant2 uint) []model.Cluster {
	t.Helper()
	clusters := []model.Cluster{
		{Name: "api-t1-default", Url: "https://api-t1-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant1},
		{Name: "api-t1-dev", Url: "https://api-t1-dev", AuthType: "token", Env: "dev", Status: "healthy", IsDefault: false, TenantID: &tenant1},
		{Name: "api-t2-default", Url: "https://api-t2-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant2},
	}
	if err := db.Create(&clusters).Error; err != nil {
		t.Fatalf("seed clusters failed: %v", err)
	}
	return clusters
}

func TestResolveClusterNameTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupK8sAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(11)
	tenant2 := uint(22)
	clusters := seedK8sTenantClusters(t, db, tenant1, tenant2)

	// clusterName 缺省时，按租户回退到当前租户默认集群。
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/pod/list", nil)
		c.Set("tenantID", tenant1)

		name, err := resolveClusterName(c)
		if err != nil {
			t.Fatalf("resolve default cluster failed: %v", err)
		}
		if name != clusters[0].Name {
			t.Fatalf("expected tenant1 default name=%s, got %s", clusters[0].Name, name)
		}
	}

	// 传入跨租户 clusterName 时必须拒绝。
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodGet,
			"/api/v1/k8s/pod/list?clusterName=api-t2-default",
			nil,
		)
		c.Set("tenantID", tenant1)

		if _, err := resolveClusterName(c); err == nil || !strings.Contains(err.Error(), "不存在或无权限") {
			t.Fatalf("expected tenant isolation error, got %v", err)
		}
	}

	// resolveListClusterName 传入空字符串时同样应按租户回退默认集群。
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/k8s/deployment/list", nil)
		c.Set("tenantID", tenant1)

		name, err := resolveListClusterName(c, "")
		if err != nil {
			t.Fatalf("resolve list default cluster failed: %v", err)
		}
		if name != clusters[0].Name {
			t.Fatalf("expected list fallback default name=%s, got %s", clusters[0].Name, name)
		}
	}
}

func TestClusterDefaultTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupK8sAPITestDB(t)
	SetK8sDB(db, nil)

	tenant1 := uint(101)
	tenant2 := uint(202)
	clusters := seedK8sTenantClusters(t, db, tenant1, tenant2)

	requestAndAssert := func(tenantID uint, expectedID uint) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/k8s/cluster/default", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("tenantID", tenantID)

		ClusterDefault(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
		var resp struct {
			Code int `json:"code"`
			Data struct {
				ID uint `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if resp.Code != 200 {
			t.Fatalf("expected code=200, got %d", resp.Code)
		}
		if resp.Data.ID != expectedID {
			t.Fatalf("expected default id=%d, got %d", expectedID, resp.Data.ID)
		}
	}

	requestAndAssert(tenant1, clusters[0].ID)
	requestAndAssert(tenant2, clusters[2].ID)
}
