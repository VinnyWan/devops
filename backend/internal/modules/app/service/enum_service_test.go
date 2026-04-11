package service

import (
	"testing"

	"devops-platform/internal/modules/app/model"
	"devops-platform/internal/modules/app/repository"
)

func TestEnumServiceDeleteEnumChecksTenantUsage(t *testing.T) {
	repo := repository.NewAppRepo()
	appSvc := NewAppServiceWithRepo(repo)
	enumSvc := NewEnumServiceWithRepo(repo)

	_, err := appSvc.SaveAppConfigInTenant(1, model.AppConfig{
		AppID:        1,
		Name:         "tenant-app",
		AppState:     "pending",
		Status:       model.StatusRunning,
		InstanceType: model.InstanceTypeContainer,
		Language:     model.LanguageGo,
	})
	if err != nil {
		t.Fatalf("save tenant app config failed: %v", err)
	}

	if err := enumSvc.DeleteEnum(1); err == nil || err.Error() != "该枚举值正在被使用，无法删除" {
		t.Fatalf("expected enum usage error, got %v", err)
	}
}
