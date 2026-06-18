package service

import (
	"testing"

	"devops-platform/internal/modules/cicd/model"
)

func TestCICDServiceSaveConfig_NilPointer(t *testing.T) {
	svc := NewCICDService(nil)

	// Verify that nil config panics (expected behavior - caller must validate)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for nil config")
		}
	}()
	svc.SaveConfig(nil)
}

func TestCICDServiceTriggerBuild_EmptyJobName(t *testing.T) {
	svc := NewCICDService(nil)

	err := svc.TriggerBuild(1, "")
	if err == nil {
		t.Fatalf("expected error for empty job name")
	}
}

func TestCICDServiceTestConnection_EmptyURL(t *testing.T) {
	svc := NewCICDService(nil)

	err := svc.TestConnection("", "admin", "token")
	if err == nil {
		t.Fatalf("expected error for empty URL")
	}
}

func TestCICDServiceSavePipeline_EmptyName(t *testing.T) {
	svc := NewCICDService(nil)

	err := svc.SavePipeline(&model.Pipeline{Name: ""})
	if err == nil {
		t.Fatalf("expected error for empty pipeline name")
	}
}

func TestCICDServiceNewService(t *testing.T) {
	svc := NewCICDService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.repo == nil {
		t.Fatal("expected non-nil repository")
	}
}
