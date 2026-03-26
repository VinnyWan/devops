package repository

import (
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/app/model"
)

type AppRepo struct {
	mu               sync.RWMutex
	apps             []model.Application
	templates        []model.AppTemplate
	deployments      []model.ApplicationDeployment
	versions         []model.ApplicationVersion
	appConfigs       []model.AppConfig
	buildConfigs     []model.BuildConfig
	deployConfigs    []model.DeployConfig
	techStackConfigs []model.TechStackConfig
	nextTemplateID   uint
	nextDeploymentID uint
	nextVersionID    uint
	nextAppConfigID  uint
	nextBuildID      uint
	nextDeployID     uint
	nextTechStackID  uint
}

func NewAppRepo() *AppRepo {
	now := time.Now()
	apps := []model.Application{
		{ID: 1, Name: "payments", Namespace: "payments", Status: "running", CreatedAt: now.Add(-90 * 24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: 2, Name: "gateway", Namespace: "gateway", Status: "running", CreatedAt: now.Add(-120 * 24 * time.Hour), UpdatedAt: now.Add(-4 * time.Hour)},
	}
	templates := []model.AppTemplate{
		{
			ID:          1,
			Name:        "payments-helm",
			Type:        "helm",
			Description: "payments 标准 Helm 模板",
			Environment: []string{"dev", "staging", "prod"},
			Variables: map[string]string{
				"replicas": "2",
				"cpu":      "500m",
				"memory":   "512Mi",
			},
			CreatedAt: now.Add(-20 * 24 * time.Hour),
			UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:          2,
			Name:        "gateway-kustomize",
			Type:        "kustomize",
			Description: "gateway 环境分层模板",
			Environment: []string{"staging", "prod"},
			Variables: map[string]string{
				"replicas": "3",
				"cpu":      "1000m",
				"memory":   "1Gi",
			},
			CreatedAt: now.Add(-14 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * time.Hour),
		},
	}
	versions := []model.ApplicationVersion{
		{ID: 1, AppID: 1, Version: "v1.8.1", Cluster: "cluster-prod", Environment: "prod", Image: "registry.example.com/payments:v1.8.1", Status: "running", Operator: "release-bot", CreatedAt: now.Add(-72 * time.Hour)},
		{ID: 2, AppID: 1, Version: "v1.8.2", Cluster: "cluster-prod", Environment: "prod", Image: "registry.example.com/payments:v1.8.2", Status: "running", Operator: "release-bot", CreatedAt: now.Add(-48 * time.Hour)},
		{ID: 3, AppID: 1, Version: "v1.9.0", Cluster: "cluster-prod", Environment: "prod", Image: "registry.example.com/payments:v1.9.0", Status: "running", Operator: "release-bot", CreatedAt: now.Add(-10 * time.Hour)},
		{ID: 4, AppID: 2, Version: "v2.3.0", Cluster: "cluster-staging", Environment: "staging", Image: "registry.example.com/gateway:v2.3.0", Status: "running", Operator: "release-bot", CreatedAt: now.Add(-24 * time.Hour)},
	}

	// ========== 配置测试数据 ==========
	appConfigs := []model.AppConfig{
		{
			ID:          1,
			AppID:       1,
			Name:        "payments",
			Owner:       "张三",
			Developers:  "李四,王五",
			Testers:     "赵六",
			GitAddress:  "https://github.com/example/payments",
			AppState:    model.AppStateRunning,
			Language:    model.LanguageJava,
			Description: "支付服务应用",
			Domain:      "payments.example.com",
			CreatedAt:   now.Add(-30 * 24 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:          2,
			AppID:       2,
			Name:        "gateway",
			Owner:       "李四",
			Developers:  "张三,王五",
			Testers:     "赵六",
			GitAddress:  "https://github.com/example/gateway",
			AppState:    model.AppStateRunning,
			Language:    model.LanguageGo,
			Description: "API网关服务",
			Domain:      "gateway.example.com",
			CreatedAt:   now.Add(-60 * 24 * time.Hour),
			UpdatedAt:   now.Add(-4 * time.Hour),
		},
	}

	buildConfigs := []model.BuildConfig{
		{
			ID:           1,
			AppID:        1,
			BuildEnv:     model.BuildEnvProduction,
			BuildTool:    "maven",
			BuildConfig:  "-DskipTests -Pprod",
			CustomConfig: "MAVEN_OPTS=-Xmx2g",
			Dockerfile: `FROM openjdk:17 AS builder
WORKDIR /build
COPY . .
RUN ["mvn", "clean", "package", "-DskipTests"]
FROM openjdk:17-jre
WORKDIR /app
COPY --from=builder /build/target/*.jar app.jar
ENTRYPOINT ["java", "-jar", "app.jar"]`,
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:           2,
			AppID:        2,
			BuildEnv:     model.BuildEnvStaging,
			BuildTool:    "go",
			BuildConfig:  "-ldflags '-s -w'",
			CustomConfig: "CGO_ENABLED=0",
			Dockerfile: `FROM golang:1.21 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app .
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT ["./app"]`,
			CreatedAt: now.Add(-60 * 24 * time.Hour),
			UpdatedAt: now.Add(-4 * time.Hour),
		},
	}

	deployConfigs := []model.DeployConfig{
		{
			ID:            1,
			AppID:         1,
			ServicePort:   8080,
			CPURequest:    "500m",
			CPULimit:      "1",
			MemoryRequest: "512Mi",
			MemoryLimit:   "1Gi",
			Environment:   model.EnvironmentProd,
			EnvVars:       "DB_HOST=mysql.prod.svc\nREDIS_HOST=redis.prod.svc",
			CreatedAt:     now.Add(-30 * 24 * time.Hour),
			UpdatedAt:     now.Add(-2 * time.Hour),
		},
		{
			ID:            2,
			AppID:         2,
			ServicePort:   8080,
			CPURequest:    "1000m",
			CPULimit:      "2",
			MemoryRequest: "1Gi",
			MemoryLimit:   "2Gi",
			Environment:   model.EnvironmentStaging,
			EnvVars:       "DB_HOST=mysql.staging.svc\nREDIS_HOST=redis.staging.svc",
			CreatedAt:     now.Add(-60 * 24 * time.Hour),
			UpdatedAt:     now.Add(-4 * time.Hour),
		},
	}

	techStackConfigs := []model.TechStackConfig{
		{
			ID:           1,
			AppID:        1,
			Name:         "Java 17",
			Language:     model.LanguageJava,
			Version:      "17",
			BaseImage:    "openjdk:17",
			BuildImage:   "maven:3.9",
			RuntimeImage: "openjdk:17-jre",
			CreatedAt:    now.Add(-30 * 24 * time.Hour),
			UpdatedAt:    now.Add(-2 * time.Hour),
		},
		{
			ID:           2,
			AppID:        2,
			Name:         "Go 1.21",
			Language:     model.LanguageGo,
			Version:      "1.21",
			BaseImage:    "golang:1.21",
			BuildImage:   "golang:1.21",
			RuntimeImage: "golang:1.21-alpine",
			CreatedAt:    now.Add(-60 * 24 * time.Hour),
			UpdatedAt:    now.Add(-4 * time.Hour),
		},
	}

	return &AppRepo{
		apps:             apps,
		templates:        templates,
		versions:         versions,
		appConfigs:       appConfigs,
		buildConfigs:     buildConfigs,
		deployConfigs:    deployConfigs,
		techStackConfigs: techStackConfigs,
		nextTemplateID:   3,
		nextDeploymentID: 1,
		nextVersionID:    5,
		nextAppConfigID:  3,
		nextBuildID:      3,
		nextDeployID:     3,
		nextTechStackID:  3,
	}
}

func (r *AppRepo) List() []model.Application {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Application(nil), r.apps...)
}

func (r *AppRepo) ListTemplates() []model.AppTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.AppTemplate, 0, len(r.templates))
	for _, item := range r.templates {
		items = append(items, cloneTemplate(item))
	}
	return items
}

func (r *AppRepo) SaveTemplate(template model.AppTemplate) model.AppTemplate {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	template = cloneTemplate(template)
	template.UpdatedAt = now
	for i, item := range r.templates {
		if item.ID != template.ID || template.ID == 0 {
			continue
		}
		template.CreatedAt = item.CreatedAt
		r.templates[i] = cloneTemplate(template)
		return cloneTemplate(r.templates[i])
	}
	if template.ID == 0 {
		template.ID = r.nextTemplateID
		r.nextTemplateID++
	}
	template.CreatedAt = now
	r.templates = append(r.templates, cloneTemplate(template))
	return cloneTemplate(template)
}

func (r *AppRepo) FindAppByID(appID uint) (model.Application, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.apps {
		if item.ID == appID {
			return item, true
		}
	}
	return model.Application{}, false
}

func (r *AppRepo) FindTemplateByID(templateID uint) (model.AppTemplate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.templates {
		if item.ID == templateID {
			return cloneTemplate(item), true
		}
	}
	return model.AppTemplate{}, false
}

func (r *AppRepo) CreateDeployment(deployment model.ApplicationDeployment) model.ApplicationDeployment {
	r.mu.Lock()
	defer r.mu.Unlock()
	deployment = cloneDeployment(deployment)
	if deployment.ID == 0 {
		deployment.ID = r.nextDeploymentID
		r.nextDeploymentID++
	}
	if deployment.CreatedAt.IsZero() {
		deployment.CreatedAt = time.Now()
	}
	r.deployments = append([]model.ApplicationDeployment{deployment}, r.deployments...)
	return cloneDeployment(deployment)
}

func (r *AppRepo) ListDeployments(appID uint, environment string) []model.ApplicationDeployment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	environment = strings.TrimSpace(strings.ToLower(environment))
	items := make([]model.ApplicationDeployment, 0, len(r.deployments))
	for _, item := range r.deployments {
		if appID > 0 && item.AppID != appID {
			continue
		}
		if environment != "" && strings.ToLower(item.Environment) != environment {
			continue
		}
		items = append(items, cloneDeployment(item))
	}
	return items
}

func (r *AppRepo) CreateVersion(version model.ApplicationVersion) model.ApplicationVersion {
	r.mu.Lock()
	defer r.mu.Unlock()
	if version.ID == 0 {
		version.ID = r.nextVersionID
		r.nextVersionID++
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now()
	}
	r.versions = append([]model.ApplicationVersion{version}, r.versions...)
	return version
}

func (r *AppRepo) ListVersions(appID uint) []model.ApplicationVersion {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.ApplicationVersion, 0, len(r.versions))
	for _, item := range r.versions {
		if appID > 0 && item.AppID != appID {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *AppRepo) FindVersion(appID uint, version string) (model.ApplicationVersion, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	target := strings.TrimSpace(strings.ToLower(version))
	for _, item := range r.versions {
		if item.AppID == appID && strings.ToLower(item.Version) == target {
			return item, true
		}
	}
	return model.ApplicationVersion{}, false
}

func cloneTemplate(in model.AppTemplate) model.AppTemplate {
	env := append([]string(nil), in.Environment...)
	vars := make(map[string]string, len(in.Variables))
	for k, v := range in.Variables {
		vars[k] = v
	}
	in.Environment = env
	in.Variables = vars
	return in
}

func cloneDeployment(in model.ApplicationDeployment) model.ApplicationDeployment {
	vars := make(map[string]string, len(in.Variables))
	for k, v := range in.Variables {
		vars[k] = v
	}
	in.Variables = vars
	return in
}

// ========== 应用配置相关方法 ==========

func (r *AppRepo) GetAppConfig(appID uint) (model.AppConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.appConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.AppConfig{}, false
}

func (r *AppRepo) SaveAppConfig(config model.AppConfig) model.AppConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.appConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			r.appConfigs[i] = config
			return config
		}
	}
	config.ID = r.nextAppConfigID
	r.nextAppConfigID++
	config.CreatedAt = now
	config.UpdatedAt = now
	r.appConfigs = append(r.appConfigs, config)
	return config
}

// ========== 构建配置相关方法 ==========

func (r *AppRepo) GetBuildConfig(appID uint) (model.BuildConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.buildConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.BuildConfig{}, false
}

func (r *AppRepo) SaveBuildConfig(config model.BuildConfig) model.BuildConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.buildConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			r.buildConfigs[i] = config
			return config
		}
	}
	config.ID = r.nextBuildID
	r.nextBuildID++
	config.CreatedAt = now
	config.UpdatedAt = now
	r.buildConfigs = append(r.buildConfigs, config)
	return config
}

// ========== 部署配置相关方法 ==========

func (r *AppRepo) GetDeployConfig(appID uint) (model.DeployConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.deployConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.DeployConfig{}, false
}

func (r *AppRepo) SaveDeployConfig(config model.DeployConfig) model.DeployConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.deployConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			r.deployConfigs[i] = config
			return config
		}
	}
	config.ID = r.nextDeployID
	r.nextDeployID++
	config.CreatedAt = now
	config.UpdatedAt = now
	r.deployConfigs = append(r.deployConfigs, config)
	return config
}

// ========== 技术栈配置相关方法 ==========

func (r *AppRepo) GetTechStackConfig(appID uint) (model.TechStackConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.techStackConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.TechStackConfig{}, false
}

func (r *AppRepo) SaveTechStackConfig(config model.TechStackConfig) model.TechStackConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.techStackConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			r.techStackConfigs[i] = config
			return config
		}
	}
	config.ID = r.nextTechStackID
	r.nextTechStackID++
	config.CreatedAt = now
	config.UpdatedAt = now
	r.techStackConfigs = append(r.techStackConfigs, config)
	return config
}

// ========== 删除配置相关方法 ==========

func (r *AppRepo) DeleteAppConfig(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.appConfigs {
		if item.AppID == appID {
			r.appConfigs = append(r.appConfigs[:i], r.appConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteBuildConfig(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.buildConfigs {
		if item.AppID == appID {
			r.buildConfigs = append(r.buildConfigs[:i], r.buildConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteDeployConfig(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.deployConfigs {
		if item.AppID == appID {
			r.deployConfigs = append(r.deployConfigs[:i], r.deployConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteTechStackConfig(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.techStackConfigs {
		if item.AppID == appID {
			r.techStackConfigs = append(r.techStackConfigs[:i], r.techStackConfigs[i+1:]...)
			return true
		}
	}
	return false
}
