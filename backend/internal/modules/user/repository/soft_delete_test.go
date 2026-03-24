package repository

import (
	"errors"
	"testing"
	"time"

	"devops-platform/internal/modules/user/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepoDeleteSoftDelete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:user_repo_soft_delete?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Permission{}, &model.Role{}, &model.Department{}, &model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	user := model.User{
		Username: "soft-delete-user",
		Password: "hashed",
		Email:    "soft-delete-user@example.com",
		Name:     "soft-delete-user",
		AuthType: model.AuthTypeLocal,
		Status:   "active",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewUserRepo(db)
	if err := repo.Delete(user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	_, err = repo.GetByID(user.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
	var raw model.User
	if err := db.Unscoped().First(&raw, user.ID).Error; err != nil {
		t.Fatalf("unscoped query: %v", err)
	}
	if !raw.DeletedAt.Valid {
		t.Fatalf("expected deleted_at set")
	}
}

func TestAuditRepoCleanExpiredSoftDelete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:audit_repo_soft_delete?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	now := time.Now().UTC()
	expired := model.AuditLog{
		UserID:        1,
		Username:      "u1",
		Operation:     "op1",
		Method:        "POST",
		Path:          "/a",
		RetentionDays: 1,
		CreatedAt:     now.AddDate(0, 0, -3),
	}
	active := model.AuditLog{
		UserID:        2,
		Username:      "u2",
		Operation:     "op2",
		Method:        "POST",
		Path:          "/b",
		RetentionDays: 7,
		CreatedAt:     now.AddDate(0, 0, -1),
	}
	if err := db.Create(&expired).Error; err != nil {
		t.Fatalf("create expired: %v", err)
	}
	if err := db.Create(&active).Error; err != nil {
		t.Fatalf("create active: %v", err)
	}
	repo := NewAuditRepo(db)
	affected, err := repo.CleanExpired(now)
	if err != nil {
		t.Fatalf("clean expired: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected affected 1, got %d", affected)
	}
	var visible int64
	if err := db.Model(&model.AuditLog{}).Count(&visible).Error; err != nil {
		t.Fatalf("count visible: %v", err)
	}
	if visible != 1 {
		t.Fatalf("expected visible 1, got %d", visible)
	}
	var total int64
	if err := db.Unscoped().Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		t.Fatalf("count unscoped: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected unscoped total 2, got %d", total)
	}
}
