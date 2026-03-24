package service

import (
	"errors"
	"testing"

	"devops-platform/internal/modules/harbor/model"
	"devops-platform/internal/pkg/obserr"
)

func TestHarborServiceListProjects_FilterKeyword(t *testing.T) {
	svc := NewHarborService()
	result, err := svc.ListProjects("pay")
	if err != nil {
		t.Fatalf("list projects failed: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected one project matched keyword, got %d", result.Total)
	}
	if result.Items[0].Name != "payments" {
		t.Fatalf("expected payments project, got %s", result.Items[0].Name)
	}

	shortKeyword, err := svc.ListProjects("pa")
	if err != nil {
		t.Fatalf("list projects with short keyword failed: %v", err)
	}
	if shortKeyword.Total != 3 {
		t.Fatalf("expected short keyword ignored and keep all, got %d", shortKeyword.Total)
	}
}

func TestHarborServiceListImages_FilterByProjectAndRepo(t *testing.T) {
	svc := NewHarborService()
	result, err := svc.ListImages("platform", "gateway")
	if err != nil {
		t.Fatalf("list images failed: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 platform gateway images, got %d", result.Total)
	}

	digestKeyword, err := svc.ListImages("platform", "abc001")
	if err != nil {
		t.Fatalf("list images by digest keyword failed: %v", err)
	}
	if digestKeyword.Total != 1 || digestKeyword.Items[0].Tag != "v1.9.0" {
		t.Fatalf("expected digest keyword matched v1.9.0, got %d", digestKeyword.Total)
	}
}

func TestHarborServiceSaveConfig_InvalidEndpoint(t *testing.T) {
	svc := NewHarborService()
	_, err := svc.SaveConfig(SaveHarborConfigRequest{
		Endpoint: "http://invalid-harbor",
	})
	if err == nil {
		t.Fatalf("expected invalid endpoint error")
	}
}

func TestHarborServiceSaveConfig_EmptyEndpoint(t *testing.T) {
	svc := NewHarborService()
	_, err := svc.SaveConfig(SaveHarborConfigRequest{})
	if err == nil {
		t.Fatalf("expected endpoint required error")
	}
	var observable *obserr.ObservableError
	if !errors.As(err, &observable) || observable.Code != "HARBOR_ENDPOINT_REQUIRED" {
		t.Fatalf("expected HARBOR_ENDPOINT_REQUIRED, got %v", err)
	}
}

func TestHarborServiceValidateCurrentConfig_InvalidConfig(t *testing.T) {
	svc := NewHarborService()
	svc.repo.SaveConfig(model.HarborConfig{Endpoint: "http://timeout-harbor"})
	err := svc.ValidateCurrentConfig()
	if err == nil {
		t.Fatalf("expected validate current config error")
	}
	var observable *obserr.ObservableError
	if !errors.As(err, &observable) || observable.Code != "HARBOR_CONNECT_FAILED" {
		t.Fatalf("expected HARBOR_CONNECT_FAILED, got %v", err)
	}
}

func TestHarborServiceListProjects_InvalidCurrentConfig(t *testing.T) {
	svc := NewHarborService()
	svc.repo.SaveConfig(model.HarborConfig{Endpoint: "http://invalid-harbor"})
	_, err := svc.ListProjects("")
	if err == nil {
		t.Fatalf("expected list projects failed by invalid config")
	}
}
