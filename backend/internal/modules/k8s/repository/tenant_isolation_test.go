package repository

import (
	"errors"
	"testing"

	"devops-platform/internal/modules/k8s/model"

	"gorm.io/gorm"
)

func TestClusterRepoTenantIsolation_DefaultListAndSetDefault(t *testing.T) {
	db := setupClusterKeywordSearchDB(t)
	repo := NewClusterRepo(db)

	tenant1 := uint(1)
	tenant2 := uint(2)
	clusters := []model.Cluster{
		{Name: "t1-default", Url: "https://t1-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant1},
		{Name: "t1-dev", Url: "https://t1-dev", AuthType: "token", Env: "dev", Status: "healthy", IsDefault: false, TenantID: &tenant1},
		{Name: "t2-default", Url: "https://t2-default", AuthType: "token", Env: "prod", Status: "healthy", IsDefault: true, TenantID: &tenant2},
	}
	if err := db.Create(&clusters).Error; err != nil {
		t.Fatalf("create clusters failed: %v", err)
	}

	list, total, err := repo.ListInTenant(tenant1, 1, 10, "", "")
	if err != nil {
		t.Fatalf("list tenant1 failed: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected tenant1 total=2, got total=%d len=%d", total, len(list))
	}
	for _, item := range list {
		if item.TenantID == nil || *item.TenantID != tenant1 {
			t.Fatalf("expected tenant1 only, got tenantID=%v", item.TenantID)
		}
	}

	default1, err := repo.GetDefaultInTenant(tenant1)
	if err != nil {
		t.Fatalf("get tenant1 default failed: %v", err)
	}
	if default1.Name != "t1-default" {
		t.Fatalf("unexpected tenant1 default: %s", default1.Name)
	}

	default2, err := repo.GetDefaultInTenant(tenant2)
	if err != nil {
		t.Fatalf("get tenant2 default failed: %v", err)
	}
	if default2.Name != "t2-default" {
		t.Fatalf("unexpected tenant2 default: %s", default2.Name)
	}

	if err := repo.SetDefaultInTenant(tenant1, clusters[1].ID); err != nil {
		t.Fatalf("set tenant1 default failed: %v", err)
	}

	default1After, err := repo.GetDefaultInTenant(tenant1)
	if err != nil {
		t.Fatalf("get tenant1 default after set failed: %v", err)
	}
	if default1After.ID != clusters[1].ID {
		t.Fatalf("expected tenant1 default id=%d, got %d", clusters[1].ID, default1After.ID)
	}

	default2After, err := repo.GetDefaultInTenant(tenant2)
	if err != nil {
		t.Fatalf("get tenant2 default after tenant1 set failed: %v", err)
	}
	if default2After.ID != clusters[2].ID {
		t.Fatalf("expected tenant2 default unchanged id=%d, got %d", clusters[2].ID, default2After.ID)
	}

	if _, err := repo.GetByIDInTenant(tenant1, clusters[2].ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected tenant1 cannot get tenant2 cluster, got err=%v", err)
	}
}
