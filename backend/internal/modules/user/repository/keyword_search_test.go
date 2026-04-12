package repository

import (
	"fmt"
	"strings"
	"testing"

	"devops-platform/internal/modules/user/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupKeywordSearchDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Permission{}, &model.Role{}, &model.Department{}, &model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestUserRepoListKeywordBoundaryAndSpecialChars(t *testing.T) {
	db := setupKeywordSearchDB(t)
	dept := model.Department{Name: "平台_%部"}
	if err := db.Create(&dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	users := []model.User{
		{Username: "alice_ops", Password: "x", Email: "alice@example.com", Name: "Alice", AuthType: model.AuthTypeLocal, Status: "active", PrimaryDeptID: &dept.ID},
		{Username: "bob", Password: "x", Email: "bob@example.com", Name: "Bob", AuthType: model.AuthTypeLocal, Status: "active", PrimaryDeptID: &dept.ID},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	repo := NewUserRepo(db)
	_, totalShort, err := repo.List(1, 10, "ab")
	if err != nil {
		t.Fatalf("list short keyword: %v", err)
	}
	if totalShort != 2 {
		t.Fatalf("expected short keyword ignored, got %d", totalShort)
	}

	items, totalExact, err := repo.List(1, 10, "ALICE")
	if err != nil {
		t.Fatalf("list exact keyword: %v", err)
	}
	if totalExact != 1 || items[0].Username != "alice_ops" {
		t.Fatalf("expected alice matched, got total=%d", totalExact)
	}

	_, totalSpecial, err := repo.List(1, 10, "%_o")
	if err != nil {
		t.Fatalf("list special keyword: %v", err)
	}
	if totalSpecial != 0 {
		t.Fatalf("expected special chars escaped and no wildcard expansion, got %d", totalSpecial)
	}
}

func TestRoleRepoListKeywordBoundaryAndSpecialChars(t *testing.T) {
	db := setupKeywordSearchDB(t)
	roles := []model.Role{
		{Name: "ops_%admin", DisplayName: "运维管理员", Type: "custom"},
		{Name: "viewer", DisplayName: "只读用户", Type: "custom"},
	}
	if err := db.Create(&roles).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}

	repo := NewRoleRepo(db)
	_, totalEmpty, err := repo.List(1, 10, "")
	if err != nil {
		t.Fatalf("list empty keyword: %v", err)
	}
	if totalEmpty != 2 {
		t.Fatalf("expected empty keyword keep all, got %d", totalEmpty)
	}

	items, totalDesc, err := repo.List(1, 10, "核心集群")
	if err != nil {
		t.Fatalf("list description keyword: %v", err)
	}
	if totalDesc != 1 || items[0].Name != "ops_%admin" {
		t.Fatalf("expected role matched by description, got total=%d", totalDesc)
	}

	_, totalSpecial, err := repo.List(1, 10, "ops_%")
	if err != nil {
		t.Fatalf("list special keyword: %v", err)
	}
	if totalSpecial != 1 {
		t.Fatalf("expected special chars literal match, got %d", totalSpecial)
	}
}

func TestPermissionRepoListKeywordBoundaryAndSpecialChars(t *testing.T) {
	db := setupKeywordSearchDB(t)
	perms := []model.Permission{
		{Name: "查看集群", Resource: "cluster", Action: "list"},
		{Name: "查看用户", Resource: "user", Action: "list"},
	}
	if err := db.Create(&perms).Error; err != nil {
		t.Fatalf("create permissions: %v", err)
	}

	repo := NewPermissionRepo(db)
	_, totalShort, err := repo.List(1, 10, "", "ab")
	if err != nil {
		t.Fatalf("list short keyword: %v", err)
	}
	if totalShort != 2 {
		t.Fatalf("expected short keyword ignored, got %d", totalShort)
	}

	items, totalResourceKeyword, err := repo.List(1, 10, "cluster", "集群")
	if err != nil {
		t.Fatalf("list with resource and keyword: %v", err)
	}
	if totalResourceKeyword != 1 || items[0].Resource != "cluster" {
		t.Fatalf("expected cluster permission matched, got total=%d", totalResourceKeyword)
	}

	_, totalSpecial, err := repo.List(1, 10, "", "%_集")
	if err != nil {
		t.Fatalf("list special keyword: %v", err)
	}
	if totalSpecial != 1 {
		t.Fatalf("expected special chars literal match, got %d", totalSpecial)
	}
}

func TestDepartmentRepoListKeywordBoundaryAndSpecialChars(t *testing.T) {
	db := setupKeywordSearchDB(t)
	depts := []model.Department{
		{Name: "平台_%部"},
		{Name: "安全合规中心"},
	}
	if err := db.Create(&depts).Error; err != nil {
		t.Fatalf("create departments: %v", err)
	}

	repo := NewDepartmentRepo(db)
	_, err := repo.List("  ")
	if err != nil {
		t.Fatalf("list blank keyword: %v", err)
	}

	items, err := repo.List("平台_%")
	if err != nil {
		t.Fatalf("list special keyword: %v", err)
	}
	if len(items) != 1 || items[0].Name != "平台_%部" {
		t.Fatalf("expected one department matched, got %d", len(items))
	}

	items, err = repo.List("安")
	if err != nil {
		t.Fatalf("list short keyword: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected short keyword ignored, got %d", len(items))
	}
}
