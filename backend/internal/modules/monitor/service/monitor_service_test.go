package service

import (
	"errors"
	"testing"
	"time"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/pkg/obserr"
)

func TestMonitorServiceQuery_Defaults(t *testing.T) {
	svc := NewMonitorService()
	start := time.Now().Add(-10 * time.Minute)
	end := time.Now()

	result, err := svc.Query("", start, end, "")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result.Metric != "cpu_usage" {
		t.Fatalf("expected default metric cpu_usage, got %s", result.Metric)
	}
	if result.Step != "1m" {
		t.Fatalf("expected default step 1m, got %s", result.Step)
	}
	if len(result.Series) == 0 {
		t.Fatalf("expected series data")
	}
}

func TestMonitorServiceQuery_EndBeforeStart(t *testing.T) {
	svc := NewMonitorService()
	start := time.Now()
	end := start.Add(-time.Minute)

	result, err := svc.Query("memory_usage", start, end, "30s")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if !result.End.After(result.Start) {
		t.Fatalf("expected end time after start time")
	}
	if result.Metric != "memory_usage" {
		t.Fatalf("expected metric memory_usage, got %s", result.Metric)
	}
}

func TestMonitorServiceSaveConfig_InvalidEndpoint(t *testing.T) {
	svc := NewMonitorService()
	_, err := svc.SaveConfig(SavePrometheusConfigRequest{
		Endpoint: "http://invalid-prometheus",
	})
	if err == nil {
		t.Fatalf("expected invalid endpoint error")
	}
}

func TestMonitorServiceQuery_InvalidStep(t *testing.T) {
	svc := NewMonitorService()
	_, err := svc.Query("cpu_usage", time.Now().Add(-time.Minute), time.Now(), "1x")
	if err == nil {
		t.Fatalf("expected invalid step error")
	}
	var observable *obserr.ObservableError
	if !errors.As(err, &observable) || observable.Code != "PROMETHEUS_INVALID_STEP" {
		t.Fatalf("expected PROMETHEUS_INVALID_STEP, got %v", err)
	}
}

func TestMonitorServiceSaveConfig_EmptyEndpoint(t *testing.T) {
	svc := NewMonitorService()
	_, err := svc.SaveConfig(SavePrometheusConfigRequest{})
	if err == nil {
		t.Fatalf("expected endpoint required error")
	}
	var observable *obserr.ObservableError
	if !errors.As(err, &observable) || observable.Code != "PROMETHEUS_ENDPOINT_REQUIRED" {
		t.Fatalf("expected PROMETHEUS_ENDPOINT_REQUIRED, got %v", err)
	}
}

func TestMonitorServiceValidateCurrentConfig_InvalidConfig(t *testing.T) {
	svc := NewMonitorService()
	svc.repo.SaveConfig(model.PrometheusConfig{Endpoint: "http://timeout-prometheus"})
	err := svc.ValidateCurrentConfig()
	if err == nil {
		t.Fatalf("expected validate current config error")
	}
	var observable *obserr.ObservableError
	if !errors.As(err, &observable) || observable.Code != "PROMETHEUS_CONNECT_FAILED" {
		t.Fatalf("expected PROMETHEUS_CONNECT_FAILED, got %v", err)
	}
}

func TestMonitorServiceQuery_InvalidCurrentConfig(t *testing.T) {
	svc := NewMonitorService()
	svc.repo.SaveConfig(model.PrometheusConfig{Endpoint: "http://invalid-prometheus"})
	_, err := svc.Query("cpu_usage", time.Now().Add(-time.Minute), time.Now(), "30s")
	if err == nil {
		t.Fatalf("expected query failed by invalid config")
	}
}
