package service

import "testing"

func TestAppServiceListTemplates_KeywordBoundaryAndCrossField(t *testing.T) {
	svc := NewAppService()
	shortKeyword := svc.ListTemplates("he")
	if shortKeyword.Total != 2 {
		t.Fatalf("expected short keyword ignored, got %d", shortKeyword.Total)
	}

	descKeyword := svc.ListTemplates("环境分层")
	if descKeyword.Total != 1 || descKeyword.Items[0].Name != "gateway-kustomize" {
		t.Fatalf("expected description keyword matched gateway-kustomize, got %d", descKeyword.Total)
	}
}

func TestAppServiceDeployAndVersions(t *testing.T) {
	svc := NewAppService()
	deployment, err := svc.Deploy(DeployRequest{
		AppID:       1,
		TemplateID:  1,
		Cluster:     "cluster-staging",
		Environment: "staging",
		Version:     "v1.9.1",
		Operator:    "tester",
	})
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}
	if deployment.ID == 0 {
		t.Fatalf("expected deployment id generated")
	}
	versions := svc.ListVersions(1, 5)
	if versions.Total == 0 {
		t.Fatalf("expected versions after deployment")
	}
	if versions.Items[0].Version != "v1.9.1" {
		t.Fatalf("expected latest version v1.9.1, got %s", versions.Items[0].Version)
	}
}

func TestAppServiceDeploy_Errors(t *testing.T) {
	svc := NewAppService()
	tests := []struct {
		name    string
		req     DeployRequest
		wantErr string
	}{
		{
			name: "app not found",
			req: DeployRequest{
				AppID:       999,
				TemplateID:  1,
				Environment: "staging",
			},
			wantErr: "应用不存在",
		},
		{
			name: "template not found",
			req: DeployRequest{
				AppID:       1,
				TemplateID:  999,
				Environment: "staging",
			},
			wantErr: "模板不存在",
		},
		{
			name: "environment unsupported",
			req: DeployRequest{
				AppID:       1,
				TemplateID:  1,
				Environment: "uat",
			},
			wantErr: "模板不支持目标环境",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Deploy(tt.req)
			if err == nil {
				t.Fatalf("expected error %s", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected error %s, got %s", tt.wantErr, err.Error())
			}
		})
	}
}

func TestAppServiceRollback(t *testing.T) {
	svc := NewAppService()
	version, err := svc.Rollback(RollbackRequest{
		AppID:    1,
		Target:   "v1.8.2",
		Operator: "tester",
	})
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if version.Status != "rolled_back" {
		t.Fatalf("expected rolled_back status, got %s", version.Status)
	}
	if version.Version != "v1.8.2" {
		t.Fatalf("expected rolled back version v1.8.2, got %s", version.Version)
	}
}

func TestAppServiceRollback_Errors(t *testing.T) {
	svc := NewAppService()
	tests := []struct {
		name    string
		req     RollbackRequest
		wantErr string
	}{
		{
			name: "app not found",
			req: RollbackRequest{
				AppID:  999,
				Target: "v1.8.2",
			},
			wantErr: "应用不存在",
		},
		{
			name: "empty target",
			req: RollbackRequest{
				AppID: 1,
			},
			wantErr: "目标版本不能为空",
		},
		{
			name: "target not found",
			req: RollbackRequest{
				AppID:  1,
				Target: "v9.9.9",
			},
			wantErr: "目标版本不存在",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Rollback(tt.req)
			if err == nil {
				t.Fatalf("expected error %s", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected error %s, got %s", tt.wantErr, err.Error())
			}
		})
	}
}

func TestAppServiceQueryTopology(t *testing.T) {
	svc := NewAppService()
	topology, err := svc.QueryTopology(1, "prod")
	if err != nil {
		t.Fatalf("query topology failed: %v", err)
	}
	if len(topology.Nodes) == 0 {
		t.Fatalf("expected topology nodes")
	}
	if len(topology.Edges) == 0 {
		t.Fatalf("expected topology edges")
	}
}

func TestAppServiceQueryTopology_AppNotFound(t *testing.T) {
	svc := NewAppService()
	_, err := svc.QueryTopology(999, "prod")
	if err == nil {
		t.Fatalf("expected app not found error")
	}
	if err.Error() != "应用不存在" {
		t.Fatalf("expected 应用不存在, got %s", err.Error())
	}
}
