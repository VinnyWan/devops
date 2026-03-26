package service

import (
	"testing"

	"devops-platform/internal/modules/app/model"
)

// TestAppConfigService_ValidateAppConfig 测试应用配置验证
func TestAppConfigService_ValidateAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  model.AppConfig
		wantErr string
	}{
		{
			name: "empty name",
			config: model.AppConfig{
				Name: "",
			},
			wantErr: "应用名称不能为空",
		},
		{
			name: "valid config",
			config: model.AppConfig{
				Name:        "test-app",
				Owner:       "张三",
				Developers:  "李四,王五",
				Testers:     "赵六",
				GitAddress:  "https://github.com/test/app",
				AppState:    model.AppStateRunning,
				Language:    model.LanguageGo,
				Description: "测试应用",
				Domain:      "test.example.com",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppConfig(tt.config)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAppConfigService_ValidateBuildConfig 测试构建配置验证
func TestAppConfigService_ValidateBuildConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  model.BuildConfig
		wantErr string
	}{
		{
			name: "missing dockerfile",
			config: model.BuildConfig{
				BuildEnv:   model.BuildEnvProduction,
				BuildTool:  "maven",
				Dockerfile: "",
			},
			wantErr: "Dockerfile不能为空",
		},
		{
			name: "valid config",
			config: model.BuildConfig{
				BuildEnv:     model.BuildEnvProduction,
				BuildTool:    "maven",
				BuildConfig:  "-DskipTests",
				Dockerfile:   "FROM openjdk:11\nWORKDIR /app",
				CustomConfig: "",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBuildConfig(tt.config)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAppConfigService_ValidateDeployConfig 测试部署配置验证
func TestAppConfigService_ValidateDeployConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  model.DeployConfig
		wantErr string
	}{
		{
			name: "invalid port",
			config: model.DeployConfig{
				ServicePort: 0,
			},
			wantErr: "服务端口必须在1-65535范围内",
		},
		{
			name: "invalid cpu",
			config: model.DeployConfig{
				ServicePort: 8080,
				CPURequest:  "invalid",
			},
			wantErr: "CPU配置无效",
		},
		{
			name: "valid config",
			config: model.DeployConfig{
				ServicePort:   8080,
				CPURequest:    "500m",
				CPULimit:      "1",
				MemoryRequest: "512Mi",
				MemoryLimit:   "1Gi",
				Environment:   model.EnvironmentProd,
				EnvVars:       "KEY1=value1",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeployConfig(tt.config)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAppConfigService_ValidateTechStackConfig 测试技术栈配置验证
func TestAppConfigService_ValidateTechStackConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  model.TechStackConfig
		wantErr string
	}{
		{
			name: "empty name",
			config: model.TechStackConfig{
				Name: "",
			},
			wantErr: "技术栈名称不能为空",
		},
		{
			name: "valid java config",
			config: model.TechStackConfig{
				Name:         "Java 17",
				Language:     model.LanguageJava,
				Version:      "17",
				BaseImage:    "openjdk:17",
				BuildImage:   "maven:3.9",
				RuntimeImage: "openjdk:17-jre",
			},
			wantErr: "",
		},
		{
			name: "valid go config",
			config: model.TechStackConfig{
				Name:         "Go 1.21",
				Language:     model.LanguageGo,
				Version:      "1.21",
				BaseImage:    "golang:1.21",
				BuildImage:   "golang:1.21",
				RuntimeImage: "golang:1.21-alpine",
			},
			wantErr: "",
		},
		{
			name: "valid python config",
			config: model.TechStackConfig{
				Name:         "Python 3.11",
				Language:     model.LanguagePython,
				Version:      "3.11",
				BaseImage:    "python:3.11",
				BuildImage:   "python:3.11",
				RuntimeImage: "python:3.11-slim",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTechStackConfig(tt.config)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
