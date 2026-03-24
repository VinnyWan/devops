package repository

import (
	"testing"
	"time"

	"devops-platform/internal/modules/user/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuditRepoListWithFilters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:audit_repo_list_filters?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	logs := []model.AuditLog{
		{
			UserID:    1,
			Username:  "alice",
			Operation: "update user",
			Method:    "POST",
			Path:      "/api/v1/user/update",
			Status:    200,
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			UserID:    2,
			Username:  "bob",
			Operation: "list role",
			Method:    "GET",
			Path:      "/api/v1/role/list",
			Status:    200,
			CreatedAt: now.Add(-1 * time.Hour),
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed logs: %v", err)
	}

	userID := uint(1)
	start := now.Add(-3 * time.Hour)
	end := now.Add(-30 * time.Minute)
	repo := NewAuditRepo(db)
	items, total, err := repo.List(AuditQuery{
		UserID:    &userID,
		Operation: "update",
		Resource:  "/user/",
		StartAt:   &start,
		EndAt:     &end,
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected item len 1, got %d", len(items))
	}
	if items[0].Username != "alice" {
		t.Fatalf("expected alice, got %s", items[0].Username)
	}
}

func TestAuditRepoListForExportRespectsLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:audit_repo_export_limit?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		entry := model.AuditLog{
			UserID:    1,
			Username:  "auditor",
			Operation: "op",
			Method:    "GET",
			Path:      "/api/v1/audit/list",
			Status:    200,
			CreatedAt: now.Add(time.Duration(-i) * time.Minute),
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("seed entry: %v", err)
		}
	}

	repo := NewAuditRepo(db)
	items, err := repo.ListForExport(AuditQuery{Username: "auditor"}, 2)
	if err != nil {
		t.Fatalf("list for export: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestAuditRepoListKeywordBoundaryAndCrossField(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:audit_repo_keyword?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	logs := []model.AuditLog{
		{
			UserID:       1,
			Username:     "alice",
			Operation:    "sync cluster",
			Method:       "GET",
			Path:         "/api/v1/k8s/cluster/list",
			Result:       `{"status":"healthy"}`,
			ErrorMessage: "",
			IP:           "10.0.0.1",
			Status:       200,
		},
		{
			UserID:       2,
			Username:     "bob",
			Operation:    "update role",
			Method:       "POST",
			Path:         "/api/v1/role/update",
			Result:       `{"status":"ok"}`,
			ErrorMessage: "permission denied",
			IP:           "10.0.0.2",
			Status:       403,
		},
	}
	if err := db.Create(&logs).Error; err != nil {
		t.Fatalf("seed logs: %v", err)
	}

	repo := NewAuditRepo(db)
	_, totalShort, err := repo.List(AuditQuery{Keyword: "ab", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list short keyword: %v", err)
	}
	if totalShort != 2 {
		t.Fatalf("expected short keyword ignored, got %d", totalShort)
	}

	items, totalPath, err := repo.List(AuditQuery{Keyword: "cluster", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list path keyword: %v", err)
	}
	if totalPath != 1 || items[0].Username != "alice" {
		t.Fatalf("expected cluster path matched alice, got total=%d", totalPath)
	}

	items, totalError, err := repo.List(AuditQuery{Keyword: "permission", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list error keyword: %v", err)
	}
	if totalError != 1 || items[0].Username != "bob" {
		t.Fatalf("expected error_message matched bob, got total=%d", totalError)
	}
}
