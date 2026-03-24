package repository

import (
	"fmt"
	"strings"
	"testing"

	"devops-platform/internal/modules/k8s/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClusterKeywordSearchDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Cluster{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestClusterRepoListKeywordBoundaryAndSpecialChars(t *testing.T) {
	db := setupClusterKeywordSearchDB(t)
	clusters := []model.Cluster{
		{
			Name:     "prod-ops_%cluster",
			Url:      "https://prod.example.com",
			AuthType: "token",
			Status:   "healthy",
			Remark:   "核心%_集群",
			Labels:   `{"team":"ops"}`,
			Env:      "prod",
		},
		{
			Name:     "dev-cluster",
			Url:      "https://dev.example.com",
			AuthType: "token",
			Status:   "healthy",
			Remark:   "开发环境",
			Labels:   `{"team":"dev"}`,
			Env:      "dev",
		},
	}
	if err := db.Create(&clusters).Error; err != nil {
		t.Fatalf("create clusters: %v", err)
	}

	repo := NewClusterRepo(db)
	_, totalShort, err := repo.List(1, 10, "", "ab")
	if err != nil {
		t.Fatalf("list short keyword: %v", err)
	}
	if totalShort != 2 {
		t.Fatalf("expected short keyword ignored, got %d", totalShort)
	}

	items, totalSpecial, err := repo.List(1, 10, "", "ops_%")
	if err != nil {
		t.Fatalf("list special keyword: %v", err)
	}
	if totalSpecial != 1 || items[0].Name != "prod-ops_%cluster" {
		t.Fatalf("expected one cluster literal matched, got total=%d", totalSpecial)
	}

	items, totalEnv, err := repo.List(1, 10, "prod", "")
	if err != nil {
		t.Fatalf("list env only: %v", err)
	}
	if totalEnv != 1 || items[0].Env != "prod" {
		t.Fatalf("expected env filter keep prod only, got total=%d", totalEnv)
	}

	_, totalNone, err := repo.List(1, 10, "prod", "dev")
	if err != nil {
		t.Fatalf("list env and keyword: %v", err)
	}
	if totalNone != 0 {
		t.Fatalf("expected env and keyword conjunction no match, got %d", totalNone)
	}
}
