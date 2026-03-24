package service

import "testing"

func TestCICDServiceListPipelineStatus_Filter(t *testing.T) {
	svc := NewCICDService()
	result := svc.ListPipelineStatus("running", "")
	if result.Total != 1 {
		t.Fatalf("expected 1 running pipeline, got %d", result.Total)
	}
	if result.Items[0].Status != "running" {
		t.Fatalf("expected running status, got %s", result.Items[0].Status)
	}

	shortKeyword := svc.ListPipelineStatus("", "ma")
	if shortKeyword.Total != 3 {
		t.Fatalf("expected short keyword ignored, got %d", shortKeyword.Total)
	}

	statusKeyword := svc.ListPipelineStatus("", "success")
	if statusKeyword.Total != 1 || statusKeyword.Items[0].Name != "gateway-release" {
		t.Fatalf("expected status field keyword matched gateway-release, got %d", statusKeyword.Total)
	}
}

func TestCICDServiceListPipelineLogs_FilterAndLimit(t *testing.T) {
	svc := NewCICDService()
	result := svc.ListPipelineLogs(1, "deploy", 1)
	if result.Total != 1 {
		t.Fatalf("expected 1 deploy log, got %d", result.Total)
	}
	if result.Items[0].Stage != "deploy" {
		t.Fatalf("expected deploy stage, got %s", result.Items[0].Stage)
	}
}

func TestCICDServiceTriggerPipeline(t *testing.T) {
	svc := NewCICDService()
	run, err := svc.TriggerPipeline(TriggerPipelineRequest{
		PipelineID:  1,
		TemplateID:  1,
		Branch:      "release/test",
		Environment: "prod",
		TriggerType: "manual",
		Operator:    "tester",
		Parameters: map[string]string{
			"strategy": "blue-green",
		},
	})
	if err != nil {
		t.Fatalf("trigger pipeline failed: %v", err)
	}
	if run.ID == 0 {
		t.Fatalf("expected run id generated")
	}
	if run.Environment != "prod" {
		t.Fatalf("expected environment prod, got %s", run.Environment)
	}
	if len(run.Stages) == 0 {
		t.Fatalf("expected stages from template")
	}
}

func TestCICDServiceSaveTemplateValidation(t *testing.T) {
	svc := NewCICDService()
	_, err := svc.SaveTemplate(SaveTemplateRequest{
		Name: "empty-template",
	})
	if err == nil {
		t.Fatalf("expected validation error for empty stages")
	}
}

func TestCICDServiceSaveConfig_InvalidEndpoint(t *testing.T) {
	svc := NewCICDService()
	_, err := svc.SaveConfig(SaveJenkinsConfigRequest{
		Endpoint: "http://invalid-jenkins",
	})
	if err == nil {
		t.Fatalf("expected invalid endpoint error")
	}
}
