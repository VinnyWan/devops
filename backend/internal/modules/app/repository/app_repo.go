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
	nextTemplateID   uint
	nextDeploymentID uint
	nextVersionID    uint
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
	return &AppRepo{
		apps:             apps,
		templates:        templates,
		versions:         versions,
		nextTemplateID:   3,
		nextDeploymentID: 1,
		nextVersionID:    5,
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
