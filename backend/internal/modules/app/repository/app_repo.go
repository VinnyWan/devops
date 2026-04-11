package repository

import (
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/app/model"
)

type AppRepo struct {
	mu                    sync.RWMutex
	apps                  []model.Application
	templates             []model.AppTemplate
	deployments           []model.ApplicationDeployment
	versions              []model.ApplicationVersion
	appConfigs            []model.AppConfig
	buildConfigs          []model.BuildConfig
	deployConfigs         []model.DeployConfig
	techStackConfigs      []model.TechStackConfig
	containerConfigs      []model.ContainerConfig
	enums                 []model.Enum
	buildEnvs             []model.BuildEnv
	nextTemplateID        uint
	nextDeploymentID      uint
	nextVersionID         uint
	nextAppConfigID       uint
	nextBuildID           uint
	nextDeployID          uint
	nextTechStackID       uint
	nextContainerConfigID uint
	nextEnumID            uint
	nextBuildEnvID        uint
	tenants               map[uint]*tenantStore
}

type tenantStore struct {
	apps                  []model.Application
	templates             []model.AppTemplate
	deployments           []model.ApplicationDeployment
	versions              []model.ApplicationVersion
	appConfigs            []model.AppConfig
	buildConfigs          []model.BuildConfig
	deployConfigs         []model.DeployConfig
	techStackConfigs      []model.TechStackConfig
	containerConfigs      []model.ContainerConfig
	nextTemplateID        uint
	nextDeploymentID      uint
	nextVersionID         uint
	nextAppConfigID       uint
	nextBuildID           uint
	nextDeployID          uint
	nextTechStackID       uint
	nextContainerConfigID uint
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
			ID:           1,
			AppID:        1,
			Name:         "payments",
			Owner:        "张三",
			Developers:   "李四,王五",
			Testers:      "赵六",
			GitAddress:   "https://github.com/example/payments",
			AppState:     model.AppStateRunning,
			Status:       model.StatusRunning,
			InstanceType: model.InstanceTypeContainer,
			Language:     model.LanguageJava,
			Port:         8080,
			Description:  "支付服务应用",
			Domain:       "payments.example.com",
			CreatedAt:    now.Add(-30 * 24 * time.Hour),
			UpdatedAt:    now.Add(-2 * time.Hour),
		},
		{
			ID:           2,
			AppID:        2,
			Name:         "gateway",
			Owner:        "李四",
			Developers:   "张三,王五",
			Testers:      "赵六",
			GitAddress:   "https://github.com/example/gateway",
			AppState:     model.AppStateRunning,
			Status:       model.StatusRunning,
			InstanceType: model.InstanceTypeContainer,
			Language:     model.LanguageGo,
			Port:         9090,
			Description:  "API网关服务",
			Domain:       "gateway.example.com",
			CreatedAt:    now.Add(-60 * 24 * time.Hour),
			UpdatedAt:    now.Add(-4 * time.Hour),
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

	// 初始化枚举数据
	enums := []model.Enum{
		// 应用状态
		{ID: 1, EnumType: "app_status", EnumKey: "pending", EnumValue: "待上线", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 2, EnumType: "app_status", EnumKey: "online", EnumValue: "已上线", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		// 运行状态
		{ID: 3, EnumType: "run_status", EnumKey: "integration", EnumValue: "集测", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 4, EnumType: "run_status", EnumKey: "staging", EnumValue: "预发", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 5, EnumType: "run_status", EnumKey: "production", EnumValue: "生产", SortOrder: 3, IsActive: true, CreatedAt: now, UpdatedAt: now},
		// 实例类型
		{ID: 6, EnumType: "instance_type", EnumKey: "container", EnumValue: "容器", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 7, EnumType: "instance_type", EnumKey: "native", EnumValue: "原方式", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		// 开发语言
		{ID: 8, EnumType: "dev_language", EnumKey: "java", EnumValue: "Java", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 9, EnumType: "dev_language", EnumKey: "go", EnumValue: "Go", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 10, EnumType: "dev_language", EnumKey: "python", EnumValue: "Python", SortOrder: 3, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 11, EnumType: "dev_language", EnumKey: "nodejs", EnumValue: "NodeJS", SortOrder: 4, IsActive: true, CreatedAt: now, UpdatedAt: now},
		// 构建工具
		{ID: 12, EnumType: "build_tool", EnumKey: "maven", EnumValue: "Maven", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 13, EnumType: "build_tool", EnumKey: "gradle", EnumValue: "Gradle", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 14, EnumType: "build_tool", EnumKey: "go", EnumValue: "Go", SortOrder: 3, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 15, EnumType: "build_tool", EnumKey: "python", EnumValue: "Python", SortOrder: 4, IsActive: true, CreatedAt: now, UpdatedAt: now},
		// 部署环境
		{ID: 16, EnumType: "environment", EnumKey: "dev", EnumValue: "开发", SortOrder: 1, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 17, EnumType: "environment", EnumKey: "test", EnumValue: "测试", SortOrder: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 18, EnumType: "environment", EnumKey: "staging", EnumValue: "预发", SortOrder: 3, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 19, EnumType: "environment", EnumKey: "prod", EnumValue: "生产", SortOrder: 4, IsActive: true, CreatedAt: now, UpdatedAt: now},
	}

	return &AppRepo{
		apps:                  apps,
		templates:             templates,
		versions:              versions,
		appConfigs:            appConfigs,
		buildConfigs:          buildConfigs,
		deployConfigs:         deployConfigs,
		techStackConfigs:      techStackConfigs,
		containerConfigs:      []model.ContainerConfig{},
		enums:                 enums,
		buildEnvs:             []model.BuildEnv{},
		nextTemplateID:        3,
		nextDeploymentID:      1,
		nextVersionID:         5,
		nextAppConfigID:       3,
		nextBuildID:           3,
		nextDeployID:          3,
		nextTechStackID:       3,
		nextContainerConfigID: 1,
		nextEnumID:            20,
		nextBuildEnvID:        1,
		tenants:               make(map[uint]*tenantStore),
	}
}

func (r *AppRepo) ensureTenantStoreLocked(tenantID uint) *tenantStore {
	if store, ok := r.tenants[tenantID]; ok {
		return store
	}

	store := &tenantStore{
		apps:                  append([]model.Application(nil), r.apps...),
		deployments:           make([]model.ApplicationDeployment, 0, len(r.deployments)),
		versions:              append([]model.ApplicationVersion(nil), r.versions...),
		appConfigs:            append([]model.AppConfig(nil), r.appConfigs...),
		buildConfigs:          append([]model.BuildConfig(nil), r.buildConfigs...),
		deployConfigs:         append([]model.DeployConfig(nil), r.deployConfigs...),
		techStackConfigs:      append([]model.TechStackConfig(nil), r.techStackConfigs...),
		containerConfigs:      append([]model.ContainerConfig(nil), r.containerConfigs...),
		nextTemplateID:        r.nextTemplateID,
		nextDeploymentID:      r.nextDeploymentID,
		nextVersionID:         r.nextVersionID,
		nextAppConfigID:       r.nextAppConfigID,
		nextBuildID:           r.nextBuildID,
		nextDeployID:          r.nextDeployID,
		nextTechStackID:       r.nextTechStackID,
		nextContainerConfigID: r.nextContainerConfigID,
	}

	store.templates = make([]model.AppTemplate, 0, len(r.templates))
	for _, item := range r.templates {
		store.templates = append(store.templates, cloneTemplate(item))
	}
	for _, item := range r.deployments {
		store.deployments = append(store.deployments, cloneDeployment(item))
	}

	r.tenants[tenantID] = store
	return store
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

// ========== 租户隔离方法 ==========

func (r *AppRepo) ListInTenant(tenantID uint) []model.Application {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	return append([]model.Application(nil), store.apps...)
}

func (r *AppRepo) ListTemplatesInTenant(tenantID uint) []model.AppTemplate {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	items := make([]model.AppTemplate, 0, len(store.templates))
	for _, item := range store.templates {
		items = append(items, cloneTemplate(item))
	}
	return items
}

func (r *AppRepo) SaveTemplateInTenant(tenantID uint, template model.AppTemplate) model.AppTemplate {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	template = cloneTemplate(template)
	template.UpdatedAt = now
	for i, item := range store.templates {
		if item.ID != template.ID || template.ID == 0 {
			continue
		}
		template.CreatedAt = item.CreatedAt
		store.templates[i] = cloneTemplate(template)
		return cloneTemplate(store.templates[i])
	}
	if template.ID == 0 {
		template.ID = store.nextTemplateID
		store.nextTemplateID++
	}
	template.CreatedAt = now
	store.templates = append(store.templates, cloneTemplate(template))
	return cloneTemplate(template)
}

func (r *AppRepo) FindAppByIDInTenant(tenantID uint, appID uint) (model.Application, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.apps {
		if item.ID == appID {
			return item, true
		}
	}
	return model.Application{}, false
}

func (r *AppRepo) FindTemplateByIDInTenant(tenantID uint, templateID uint) (model.AppTemplate, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.templates {
		if item.ID == templateID {
			return cloneTemplate(item), true
		}
	}
	return model.AppTemplate{}, false
}

func (r *AppRepo) CreateDeploymentInTenant(tenantID uint, deployment model.ApplicationDeployment) model.ApplicationDeployment {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	deployment = cloneDeployment(deployment)
	if deployment.ID == 0 {
		deployment.ID = store.nextDeploymentID
		store.nextDeploymentID++
	}
	if deployment.CreatedAt.IsZero() {
		deployment.CreatedAt = now
	}
	store.deployments = append(store.deployments, cloneDeployment(deployment))
	return cloneDeployment(deployment)
}

func (r *AppRepo) ListDeploymentsInTenant(tenantID uint, appID uint, environment string) []model.ApplicationDeployment {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	targetEnv := strings.TrimSpace(strings.ToLower(environment))
	items := make([]model.ApplicationDeployment, 0, len(store.deployments))
	for _, item := range store.deployments {
		if appID > 0 && item.AppID != appID {
			continue
		}
		if targetEnv != "" && strings.ToLower(item.Environment) != targetEnv {
			continue
		}
		items = append(items, cloneDeployment(item))
	}
	return items
}

func (r *AppRepo) CreateVersionInTenant(tenantID uint, version model.ApplicationVersion) model.ApplicationVersion {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	if version.ID == 0 {
		version.ID = store.nextVersionID
		store.nextVersionID++
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}
	store.versions = append(store.versions, version)
	return version
}

func (r *AppRepo) ListVersionsInTenant(tenantID uint, appID uint) []model.ApplicationVersion {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	items := make([]model.ApplicationVersion, 0, len(store.versions))
	for _, item := range store.versions {
		if appID > 0 && item.AppID != appID {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (r *AppRepo) FindVersionInTenant(tenantID uint, appID uint, version string) (model.ApplicationVersion, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	target := strings.TrimSpace(strings.ToLower(version))
	for _, item := range store.versions {
		if item.AppID == appID && strings.ToLower(item.Version) == target {
			return item, true
		}
	}
	return model.ApplicationVersion{}, false
}

func (r *AppRepo) GetAppConfigInTenant(tenantID uint, appID uint) (model.AppConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.appConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.AppConfig{}, false
}

func (r *AppRepo) SaveAppConfigInTenant(tenantID uint, config model.AppConfig) model.AppConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	for i, item := range store.appConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			store.appConfigs[i] = config
			return config
		}
	}
	config.ID = store.nextAppConfigID
	store.nextAppConfigID++
	config.CreatedAt = now
	config.UpdatedAt = now
	store.appConfigs = append(store.appConfigs, config)
	return config
}

func (r *AppRepo) GetBuildConfigInTenant(tenantID uint, appID uint) (model.BuildConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.buildConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.BuildConfig{}, false
}

func (r *AppRepo) SaveBuildConfigInTenant(tenantID uint, config model.BuildConfig) model.BuildConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	for i, item := range store.buildConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			store.buildConfigs[i] = config
			return config
		}
	}
	config.ID = store.nextBuildID
	store.nextBuildID++
	config.CreatedAt = now
	config.UpdatedAt = now
	store.buildConfigs = append(store.buildConfigs, config)
	return config
}

func (r *AppRepo) GetDeployConfigInTenant(tenantID uint, appID uint, environment string) (model.DeployConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.deployConfigs {
		if item.AppID == appID && item.Environment == environment {
			return item, true
		}
	}
	return model.DeployConfig{}, false
}

func (r *AppRepo) GetDeployConfigByAppIDInTenant(tenantID uint, appID uint) (model.DeployConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.deployConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.DeployConfig{}, false
}

func (r *AppRepo) ListDeployConfigsByAppInTenant(tenantID uint, appID uint) []model.DeployConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	var result []model.DeployConfig
	for _, item := range store.deployConfigs {
		if item.AppID == appID {
			result = append(result, item)
		}
	}
	return result
}

func (r *AppRepo) SaveDeployConfigInTenant(tenantID uint, config model.DeployConfig) model.DeployConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	for i, item := range store.deployConfigs {
		if item.AppID == config.AppID && item.Environment == config.Environment {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			store.deployConfigs[i] = config
			return config
		}
	}
	config.ID = store.nextDeployID
	store.nextDeployID++
	config.CreatedAt = now
	config.UpdatedAt = now
	store.deployConfigs = append(store.deployConfigs, config)
	return config
}

func (r *AppRepo) DeleteDeployConfigByEnvInTenant(tenantID uint, appID uint, environment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, item := range store.deployConfigs {
		if item.AppID == appID && item.Environment == environment {
			store.deployConfigs = append(store.deployConfigs[:i], store.deployConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) GetTechStackConfigInTenant(tenantID uint, appID uint) (model.TechStackConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.techStackConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.TechStackConfig{}, false
}

func (r *AppRepo) SaveTechStackConfigInTenant(tenantID uint, config model.TechStackConfig) model.TechStackConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	for i, item := range store.techStackConfigs {
		if item.AppID == config.AppID {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			store.techStackConfigs[i] = config
			return config
		}
	}
	config.ID = store.nextTechStackID
	store.nextTechStackID++
	config.CreatedAt = now
	config.UpdatedAt = now
	store.techStackConfigs = append(store.techStackConfigs, config)
	return config
}

func (r *AppRepo) DeleteAppConfigInTenant(tenantID uint, appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, item := range store.appConfigs {
		if item.AppID == appID {
			store.appConfigs = append(store.appConfigs[:i], store.appConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteAppConfigCascadeInTenant(tenantID uint, appID uint) bool {
	deleted := false
	if r.DeleteAppConfigInTenant(tenantID, appID) {
		deleted = true
	}
	if r.DeleteBuildConfigInTenant(tenantID, appID) {
		deleted = true
	}
	if r.DeleteDeployConfigInTenant(tenantID, appID) {
		deleted = true
	}
	if r.DeleteTechStackConfigInTenant(tenantID, appID) {
		deleted = true
	}
	if r.DeleteContainerConfigInTenant(tenantID, appID) {
		deleted = true
	}
	return deleted
}

func (r *AppRepo) DeleteBuildConfigInTenant(tenantID uint, appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, item := range store.buildConfigs {
		if item.AppID == appID {
			store.buildConfigs = append(store.buildConfigs[:i], store.buildConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteDeployConfigInTenant(tenantID uint, appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	var newConfigs []model.DeployConfig
	for _, item := range store.deployConfigs {
		if item.AppID != appID {
			newConfigs = append(newConfigs, item)
		}
	}
	if len(newConfigs) != len(store.deployConfigs) {
		store.deployConfigs = newConfigs
		return true
	}
	return false
}

func (r *AppRepo) DeleteTechStackConfigInTenant(tenantID uint, appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, item := range store.techStackConfigs {
		if item.AppID == appID {
			store.techStackConfigs = append(store.techStackConfigs[:i], store.techStackConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) ListAppsWithFilterInTenant(tenantID uint, page, pageSize int, keyword, instanceType, status string) ([]model.AppConfig, int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)

	var filtered []model.AppConfig
	for _, config := range store.appConfigs {
		if keyword != "" {
			kw := strings.ToLower(keyword)
			if !strings.Contains(strings.ToLower(config.Name), kw) &&
				!strings.Contains(strings.ToLower(config.Owner), kw) &&
				!strings.Contains(strings.ToLower(config.Developers), kw) {
				continue
			}
		}
		if instanceType != "" && config.InstanceType != instanceType {
			continue
		}
		if status != "" && config.Status != status {
			continue
		}
		filtered = append(filtered, config)
	}

	total := int64(len(filtered))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(filtered) {
		return []model.AppConfig{}, total
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total
}

func (r *AppRepo) ToggleAppStatusInTenant(tenantID uint, appID uint, status string) (model.AppConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, config := range store.appConfigs {
		if config.AppID == appID {
			store.appConfigs[i].Status = status
			store.appConfigs[i].UpdatedAt = time.Now()
			return store.appConfigs[i], true
		}
	}
	return model.AppConfig{}, false
}

func (r *AppRepo) GetContainerConfigInTenant(tenantID uint, appID uint, environment string) (model.ContainerConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.containerConfigs {
		if item.AppID == appID && item.Environment == environment {
			return item, true
		}
	}
	return model.ContainerConfig{}, false
}

func (r *AppRepo) GetContainerConfigByAppIDInTenant(tenantID uint, appID uint) (model.ContainerConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for _, item := range store.containerConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.ContainerConfig{}, false
}

func (r *AppRepo) ListContainerConfigsByAppInTenant(tenantID uint, appID uint) []model.ContainerConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	var result []model.ContainerConfig
	for _, item := range store.containerConfigs {
		if item.AppID == appID {
			result = append(result, item)
		}
	}
	return result
}

func (r *AppRepo) SaveContainerConfigInTenant(tenantID uint, config model.ContainerConfig) model.ContainerConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	now := time.Now()
	for i, item := range store.containerConfigs {
		if item.AppID == config.AppID && item.Environment == config.Environment {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			store.containerConfigs[i] = config
			return config
		}
	}
	config.ID = store.nextContainerConfigID
	store.nextContainerConfigID++
	config.CreatedAt = now
	config.UpdatedAt = now
	store.containerConfigs = append(store.containerConfigs, config)
	return config
}

func (r *AppRepo) DeleteContainerConfigByEnvInTenant(tenantID uint, appID uint, environment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	for i, item := range store.containerConfigs {
		if item.AppID == appID && item.Environment == environment {
			store.containerConfigs = append(store.containerConfigs[:i], store.containerConfigs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *AppRepo) DeleteContainerConfigInTenant(tenantID uint, appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	store := r.ensureTenantStoreLocked(tenantID)
	var newConfigs []model.ContainerConfig
	for _, item := range store.containerConfigs {
		if item.AppID != appID {
			newConfigs = append(newConfigs, item)
		}
	}
	if len(newConfigs) != len(store.containerConfigs) {
		store.containerConfigs = newConfigs
		return true
	}
	return false
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

// GetDeployConfig 获取指定应用和环境的部署配置
func (r *AppRepo) GetDeployConfig(appID uint, environment string) (model.DeployConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.deployConfigs {
		if item.AppID == appID && item.Environment == environment {
			return item, true
		}
	}
	return model.DeployConfig{}, false
}

// GetDeployConfigByAppID 获取应用的部署配置（兼容旧接口，返回第一个）
func (r *AppRepo) GetDeployConfigByAppID(appID uint) (model.DeployConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.deployConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.DeployConfig{}, false
}

// ListDeployConfigsByApp 获取应用所有环境的部署配置
func (r *AppRepo) ListDeployConfigsByApp(appID uint) []model.DeployConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []model.DeployConfig
	for _, item := range r.deployConfigs {
		if item.AppID == appID {
			result = append(result, item)
		}
	}
	return result
}

// SaveDeployConfig 保存部署配置（按appID+environment唯一约束）
func (r *AppRepo) SaveDeployConfig(config model.DeployConfig) model.DeployConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.deployConfigs {
		if item.AppID == config.AppID && item.Environment == config.Environment {
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

// DeleteDeployConfigByEnv 删除指定环境的部署配置
func (r *AppRepo) DeleteDeployConfigByEnv(appID uint, environment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.deployConfigs {
		if item.AppID == appID && item.Environment == environment {
			r.deployConfigs = append(r.deployConfigs[:i], r.deployConfigs[i+1:]...)
			return true
		}
	}
	return false
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

// DeleteAppConfigCascade 级联删除应用及其所有关联配置
func (r *AppRepo) DeleteAppConfigCascade(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 删除应用配置
	var newAppConfigs []model.AppConfig
	for _, item := range r.appConfigs {
		if item.AppID != appID {
			newAppConfigs = append(newAppConfigs, item)
		}
	}
	r.appConfigs = newAppConfigs

	// 删除构建配置
	var newBuildConfigs []model.BuildConfig
	for _, item := range r.buildConfigs {
		if item.AppID != appID {
			newBuildConfigs = append(newBuildConfigs, item)
		}
	}
	r.buildConfigs = newBuildConfigs

	// 删除部署配置
	var newDeployConfigs []model.DeployConfig
	for _, item := range r.deployConfigs {
		if item.AppID != appID {
			newDeployConfigs = append(newDeployConfigs, item)
		}
	}
	r.deployConfigs = newDeployConfigs

	// 删除容器配置
	var newContainerConfigs []model.ContainerConfig
	for _, item := range r.containerConfigs {
		if item.AppID != appID {
			newContainerConfigs = append(newContainerConfigs, item)
		}
	}
	r.containerConfigs = newContainerConfigs

	// 删除技术栈配置
	var newTechStackConfigs []model.TechStackConfig
	for _, item := range r.techStackConfigs {
		if item.AppID != appID {
			newTechStackConfigs = append(newTechStackConfigs, item)
		}
	}
	r.techStackConfigs = newTechStackConfigs

	return true
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
	// 删除该应用的所有环境配置
	var newConfigs []model.DeployConfig
	for _, item := range r.deployConfigs {
		if item.AppID != appID {
			newConfigs = append(newConfigs, item)
		}
	}
	if len(newConfigs) != len(r.deployConfigs) {
		r.deployConfigs = newConfigs
		return true
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

// ========== 分页查询相关方法 ==========

// ListAppsWithFilter 分页查询应用列表，支持搜索和筛选
func (r *AppRepo) ListAppsWithFilter(page, pageSize int, keyword, instanceType, status string) ([]model.AppConfig, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 筛选
	var filtered []model.AppConfig
	for _, config := range r.appConfigs {
		// 关键字匹配（应用名称、负责人）
		if keyword != "" {
			kw := strings.ToLower(keyword)
			if !strings.Contains(strings.ToLower(config.Name), kw) &&
				!strings.Contains(strings.ToLower(config.Owner), kw) &&
				!strings.Contains(strings.ToLower(config.Developers), kw) {
				continue
			}
		}
		// 实例类型筛选
		if instanceType != "" && config.InstanceType != instanceType {
			continue
		}
		// 状态筛选
		if status != "" && config.Status != status {
			continue
		}
		filtered = append(filtered, config)
	}

	total := int64(len(filtered))

	// 分页
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filtered) {
		return []model.AppConfig{}, total
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total
}

// ToggleAppStatus 切换应用状态
func (r *AppRepo) ToggleAppStatus(appID uint, status string) (model.AppConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, config := range r.appConfigs {
		if config.AppID == appID {
			r.appConfigs[i].Status = status
			r.appConfigs[i].UpdatedAt = time.Now()
			return r.appConfigs[i], true
		}
	}
	return model.AppConfig{}, false
}

// ========== 容器配置相关方法 ==========

// GetContainerConfig 获取指定应用和环境的容器配置
func (r *AppRepo) GetContainerConfig(appID uint, environment string) (model.ContainerConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.containerConfigs {
		if item.AppID == appID && item.Environment == environment {
			return item, true
		}
	}
	return model.ContainerConfig{}, false
}

// GetContainerConfigByAppID 获取应用的容器配置（兼容旧接口，返回第一个）
func (r *AppRepo) GetContainerConfigByAppID(appID uint) (model.ContainerConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.containerConfigs {
		if item.AppID == appID {
			return item, true
		}
	}
	return model.ContainerConfig{}, false
}

// ListContainerConfigsByApp 获取应用所有环境的容器配置
func (r *AppRepo) ListContainerConfigsByApp(appID uint) []model.ContainerConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []model.ContainerConfig
	for _, item := range r.containerConfigs {
		if item.AppID == appID {
			result = append(result, item)
		}
	}
	return result
}

// DeleteContainerConfigByEnv 删除指定环境的容器配置
func (r *AppRepo) SaveContainerConfig(config model.ContainerConfig) model.ContainerConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for i, item := range r.containerConfigs {
		if item.AppID == config.AppID && item.Environment == config.Environment {
			config.ID = item.ID
			config.CreatedAt = item.CreatedAt
			config.UpdatedAt = now
			r.containerConfigs[i] = config
			return config
		}
	}
	config.ID = r.nextContainerConfigID
	r.nextContainerConfigID++
	config.CreatedAt = now
	config.UpdatedAt = now
	r.containerConfigs = append(r.containerConfigs, config)
	return config
}

// DeleteContainerConfigByEnv 删除指定环境的容器配置
func (r *AppRepo) DeleteContainerConfigByEnv(appID uint, environment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.containerConfigs {
		if item.AppID == appID && item.Environment == environment {
			r.containerConfigs = append(r.containerConfigs[:i], r.containerConfigs[i+1:]...)
			return true
		}
	}
	return false
}

// DeleteContainerConfig 删除应用的所有容器配置
func (r *AppRepo) DeleteContainerConfig(appID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	var newConfigs []model.ContainerConfig
	for _, item := range r.containerConfigs {
		if item.AppID != appID {
			newConfigs = append(newConfigs, item)
		}
	}
	if len(newConfigs) != len(r.containerConfigs) {
		r.containerConfigs = newConfigs
		return true
	}
	return false
}

// ========== 枚举管理相关方法 ==========

// ListEnums 获取枚举列表，支持按类型筛选
func (r *AppRepo) ListEnums(enumType string) []model.Enum {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.Enum
	for _, e := range r.enums {
		if enumType == "" || e.EnumType == enumType {
			if e.IsActive {
				result = append(result, e)
			}
		}
	}

	// 按 SortOrder 排序
	sortEnums(result)
	return result
}

// ListAllEnums 获取所有枚举（包括禁用的）
func (r *AppRepo) ListAllEnums(enumType string) []model.Enum {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.Enum
	for _, e := range r.enums {
		if enumType == "" || e.EnumType == enumType {
			result = append(result, e)
		}
	}

	sortEnums(result)
	return result
}

// GetEnumByID 根据ID获取枚举
func (r *AppRepo) GetEnumByID(id uint) (model.Enum, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.enums {
		if e.ID == id {
			return e, true
		}
	}
	return model.Enum{}, false
}

// SaveEnum 保存枚举（新增或更新）
func (r *AppRepo) SaveEnum(enum model.Enum) model.Enum {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()

	// 查找是否已存在
	for i, e := range r.enums {
		if e.ID == enum.ID && enum.ID != 0 {
			enum.CreatedAt = e.CreatedAt
			enum.UpdatedAt = now
			r.enums[i] = enum
			return enum
		}
	}

	// 新增
	if enum.ID == 0 {
		enum.ID = r.nextEnumID
		r.nextEnumID++
	}
	enum.CreatedAt = now
	enum.UpdatedAt = now
	r.enums = append(r.enums, enum)
	return enum
}

// DeleteEnum 删除枚举
func (r *AppRepo) DeleteEnum(id uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.enums {
		if e.ID == id {
			r.enums = append(r.enums[:i], r.enums[i+1:]...)
			return true
		}
	}
	return false
}

// CheckEnumUsage 检查枚举是否被使用
func (r *AppRepo) CheckEnumUsage(enumType, enumKey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if checkEnumUsageInStore(enumType, enumKey, r.appConfigs, r.buildConfigs, r.deployConfigs, r.containerConfigs) {
		return true
	}
	for _, store := range r.tenants {
		if checkEnumUsageInStore(enumType, enumKey, store.appConfigs, store.buildConfigs, store.deployConfigs, store.containerConfigs) {
			return true
		}
	}
	return false
}

func checkEnumUsageInStore(
	enumType string,
	enumKey string,
	appConfigs []model.AppConfig,
	buildConfigs []model.BuildConfig,
	deployConfigs []model.DeployConfig,
	containerConfigs []model.ContainerConfig,
) bool {
	switch enumType {
	case "app_status", "run_status", "instance_type", "dev_language":
		for _, config := range appConfigs {
			if enumType == "app_status" && config.AppState == enumKey {
				return true
			}
			if enumType == "run_status" && config.Status == enumKey {
				return true
			}
			if enumType == "instance_type" && config.InstanceType == enumKey {
				return true
			}
			if enumType == "dev_language" && config.Language == enumKey {
				return true
			}
		}
	case "build_tool":
		for _, config := range buildConfigs {
			if config.BuildTool == enumKey {
				return true
			}
		}
	case "environment":
		for _, config := range deployConfigs {
			if config.Environment == enumKey {
				return true
			}
		}
		for _, config := range containerConfigs {
			// 容器配置暂无环境字段，后续阶段添加
			_ = config
		}
	}
	return false
}

// GetEnumTypes 获取所有枚举类型
func (r *AppRepo) GetEnumTypes() []string {
	return []string{
		"app_status",
		"run_status",
		"instance_type",
		"dev_language",
		"build_tool",
		"environment",
	}
}

// sortEnums 按SortOrder排序枚举
func sortEnums(enums []model.Enum) {
	for i := 0; i < len(enums)-1; i++ {
		for j := i + 1; j < len(enums); j++ {
			if enums[i].SortOrder > enums[j].SortOrder {
				enums[i], enums[j] = enums[j], enums[i]
			}
		}
	}
}

// ========== 构建环境版本管理相关方法 ==========

// ListBuildEnvs 获取构建环境版本列表
func (r *AppRepo) ListBuildEnvs() []model.BuildEnv {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.BuildEnv, len(r.buildEnvs))
	copy(result, r.buildEnvs)
	return result
}

// GetBuildEnvByID 根据ID获取构建环境
func (r *AppRepo) GetBuildEnvByID(id uint) (model.BuildEnv, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.buildEnvs {
		if e.ID == id {
			return e, true
		}
	}
	return model.BuildEnv{}, false
}

// SaveBuildEnv 保存构建环境
func (r *AppRepo) SaveBuildEnv(env model.BuildEnv) model.BuildEnv {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()

	for i, e := range r.buildEnvs {
		if e.ID == env.ID && env.ID != 0 {
			env.CreatedAt = e.CreatedAt
			env.UpdatedAt = now
			r.buildEnvs[i] = env
			return env
		}
	}

	if env.ID == 0 {
		env.ID = r.nextBuildEnvID
		r.nextBuildEnvID++
	}
	env.CreatedAt = now
	env.UpdatedAt = now
	r.buildEnvs = append(r.buildEnvs, env)
	return env
}

// DeleteBuildEnv 删除构建环境
func (r *AppRepo) DeleteBuildEnv(id uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.buildEnvs {
		if e.ID == id {
			r.buildEnvs = append(r.buildEnvs[:i], r.buildEnvs[i+1:]...)
			return true
		}
	}
	return false
}
