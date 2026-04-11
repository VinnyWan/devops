package service

import (
	"testing"

	"devops-platform/internal/modules/k8s/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClusterServiceTenantDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:cluster_service_tenant?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	if err := db.AutoMigrate(&model.Cluster{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

func TestClusterServiceTenantIsolation_SetDefaultAndDelete(t *testing.T) {
	db := setupClusterServiceTenantDB(t)
	svc := NewClusterService(db)

	tenant1 := uint(101)
	tenant2 := uint(202)
	clusters := []model.Cluster{
		{Name: "svc-t1-default", Url: "https://svc-t1-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant1},
		{Name: "svc-t1-dev", Url: "https://svc-t1-dev", AuthType: "token", Env: "dev", Status: "healthy", IsDefault: false, TenantID: &tenant1},
		{Name: "svc-t2-default", Url: "https://svc-t2-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant2},
	}
	if err := db.Create(&clusters).Error; err != nil {
		t.Fatalf("create clusters failed: %v", err)
	}

	items, total, err := svc.ListInTenant(tenant1, 1, 10, "", "")
	if err != nil {
		t.Fatalf("list tenant1 failed: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected tenant1 total=2 len=2, got total=%d len=%d", total, len(items))
	}

	if err := svc.SetDefaultInTenant(tenant1, clusters[1].ID); err != nil {
		t.Fatalf("set tenant1 default failed: %v", err)
	}

	default1, err := svc.GetDefaultOrFirstInTenant(tenant1)
	if err != nil {
		t.Fatalf("get tenant1 default failed: %v", err)
	}
	if default1.ID != clusters[1].ID {
		t.Fatalf("expected tenant1 default id=%d, got %d", clusters[1].ID, default1.ID)
	}

	default2, err := svc.GetDefaultOrFirstInTenant(tenant2)
	if err != nil {
		t.Fatalf("get tenant2 default failed: %v", err)
	}
	if default2.ID != clusters[2].ID {
		t.Fatalf("expected tenant2 default unchanged id=%d, got %d", clusters[2].ID, default2.ID)
	}

	if err := svc.DeleteInTenant(tenant1, clusters[2].ID); err == nil || err.Error() != "集群不存在" {
		t.Fatalf("expected tenant delete isolation error, got %v", err)
	}
}
