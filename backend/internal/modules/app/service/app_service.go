package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"devops-platform/internal/modules/app/model"
	"devops-platform/internal/modules/app/repository"
	queryutil "devops-platform/internal/pkg/query"
)

type AppService struct {
	repo *repository.AppRepo
}

type ListTemplatesResponse struct {
	Total int                 `json:"total"`
	Items []model.AppTemplate `json:"items"`
}

type SaveTemplateRequest struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Environment []string          `json:"environment"`
	Variables   map[string]string `json:"variables"`
}

type DeployRequest struct {
	AppID       uint              `json:"appId"`
	TemplateID  uint              `json:"templateId"`
	Cluster     string            `json:"cluster"`
	Environment string            `json:"environment"`
	Namespace   string            `json:"namespace"`
	Version     string            `json:"version"`
	Operator    string            `json:"operator"`
	Variables   map[string]string `json:"variables"`
}

type ListDeploymentsResponse struct {
	Total int                           `json:"total"`
	Items []model.ApplicationDeployment `json:"items"`
}

type ListVersionsResponse struct {
	Total int                        `json:"total"`
	Items []model.ApplicationVersion `json:"items"`
}

type RollbackRequest struct {
	AppID       uint   `json:"appId"`
	Target      string `json:"target"`
	Cluster     string `json:"cluster"`
	Environment string `json:"environment"`
	Operator    string `json:"operator"`
}

func NewAppService() *AppService {
	return &AppService{repo: repository.NewAppRepo()}
}

func (s *AppService) List() []model.Application {
	return s.repo.List()
}

func (s *AppService) ListTemplates(keyword string) ListTemplatesResponse {
	templates := s.repo.ListTemplates()
	items := make([]model.AppTemplate, 0, len(templates))
	for _, item := range templates {
		if !queryutil.MatchKeywordAny(keyword, item.Name, item.Type, item.Description, strings.Join(item.Environment, ",")) {
			continue
		}
		items = append(items, item)
	}
	return ListTemplatesResponse{Total: len(items), Items: items}
}

func (s *AppService) SaveTemplate(req SaveTemplateRequest) (model.AppTemplate, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return model.AppTemplate{}, errors.New("模板名称不能为空")
	}
	templateType := strings.TrimSpace(strings.ToLower(req.Type))
	if templateType == "" {
		templateType = "helm"
	}
	environment := make([]string, 0, len(req.Environment))
	seen := map[string]struct{}{}
	for _, env := range req.Environment {
		v := strings.TrimSpace(strings.ToLower(env))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		environment = append(environment, v)
	}
	if len(environment) == 0 {
		environment = []string{"staging"}
	}
	variables := make(map[string]string, len(req.Variables))
	for k, v := range req.Variables {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		variables[key] = strings.TrimSpace(v)
	}
	template := s.repo.SaveTemplate(model.AppTemplate{
		ID:          req.ID,
		Name:        name,
		Type:        templateType,
		Description: strings.TrimSpace(req.Description),
		Environment: environment,
		Variables:   variables,
	})
	return template, nil
}

func (s *AppService) Deploy(req DeployRequest) (model.ApplicationDeployment, error) {
	app, ok := s.repo.FindAppByID(req.AppID)
	if !ok {
		return model.ApplicationDeployment{}, errors.New("应用不存在")
	}
	template, ok := s.repo.FindTemplateByID(req.TemplateID)
	if !ok {
		return model.ApplicationDeployment{}, errors.New("模板不存在")
	}
	cluster := strings.TrimSpace(req.Cluster)
	if cluster == "" {
		cluster = "default-cluster"
	}
	environment := strings.TrimSpace(strings.ToLower(req.Environment))
	if environment == "" {
		environment = "staging"
	}
	if !containsString(template.Environment, environment) {
		return model.ApplicationDeployment{}, errors.New("模板不支持目标环境")
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = fmt.Sprintf("v%s", time.Now().Format("20060102150405"))
	}
	namespace := strings.TrimSpace(req.Namespace)
	if namespace == "" {
		namespace = app.Namespace
	}
	operator := strings.TrimSpace(req.Operator)
	if operator == "" {
		operator = "system"
	}
	variables := mergeVariables(template.Variables, req.Variables)
	deployment := s.repo.CreateDeployment(model.ApplicationDeployment{
		AppID:        app.ID,
		AppName:      app.Name,
		TemplateID:   template.ID,
		TemplateName: template.Name,
		Cluster:      cluster,
		Environment:  environment,
		Namespace:    namespace,
		Version:      version,
		Status:       "deployed",
		Operator:     operator,
		Variables:    variables,
	})
	s.repo.CreateVersion(model.ApplicationVersion{
		AppID:       app.ID,
		Version:     version,
		Cluster:     cluster,
		Environment: environment,
		Image:       fmt.Sprintf("registry.example.com/%s:%s", app.Name, version),
		Status:      "running",
		Operator:    operator,
	})
	return deployment, nil
}

func (s *AppService) ListDeployments(appID uint, environment string, limit int) ListDeploymentsResponse {
	if limit <= 0 {
		limit = 20
	}
	items := s.repo.ListDeployments(appID, environment)
	if len(items) > limit {
		items = items[:limit]
	}
	return ListDeploymentsResponse{Total: len(items), Items: items}
}

func (s *AppService) ListVersions(appID uint, limit int) ListVersionsResponse {
	if limit <= 0 {
		limit = 20
	}
	items := s.repo.ListVersions(appID)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return ListVersionsResponse{Total: len(items), Items: items}
}

func (s *AppService) Rollback(req RollbackRequest) (model.ApplicationVersion, error) {
	app, ok := s.repo.FindAppByID(req.AppID)
	if !ok {
		return model.ApplicationVersion{}, errors.New("应用不存在")
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		return model.ApplicationVersion{}, errors.New("目标版本不能为空")
	}
	version, ok := s.repo.FindVersion(app.ID, target)
	if !ok {
		return model.ApplicationVersion{}, errors.New("目标版本不存在")
	}
	cluster := strings.TrimSpace(req.Cluster)
	if cluster == "" {
		cluster = version.Cluster
	}
	environment := strings.TrimSpace(strings.ToLower(req.Environment))
	if environment == "" {
		environment = version.Environment
	}
	operator := strings.TrimSpace(req.Operator)
	if operator == "" {
		operator = "system"
	}
	rolled := s.repo.CreateVersion(model.ApplicationVersion{
		AppID:       app.ID,
		Version:     version.Version,
		Cluster:     cluster,
		Environment: environment,
		Image:       version.Image,
		Status:      "rolled_back",
		Operator:    operator,
		CreatedAt:   time.Now(),
	})
	return rolled, nil
}

func (s *AppService) QueryTopology(appID uint, environment string) (model.ApplicationTopology, error) {
	app, ok := s.repo.FindAppByID(appID)
	if !ok {
		return model.ApplicationTopology{}, errors.New("应用不存在")
	}
	environment = strings.TrimSpace(strings.ToLower(environment))
	if environment == "" {
		environment = "staging"
	}
	nodes := []model.TopologyNode{
		{ID: "deploy-" + app.Name, Name: app.Name + "-deployment", Kind: "Deployment", Status: "Running", Cluster: "multi-cluster", Metadata: environment},
		{ID: "svc-" + app.Name, Name: app.Name + "-service", Kind: "Service", Status: "Ready", Cluster: "multi-cluster", Metadata: environment},
		{ID: "ing-" + app.Name, Name: app.Name + "-ingress", Kind: "Ingress", Status: "Ready", Cluster: "multi-cluster", Metadata: environment},
		{ID: "cfg-" + app.Name, Name: app.Name + "-config", Kind: "ConfigMap", Status: "Synced", Cluster: "multi-cluster", Metadata: environment},
	}
	edges := []model.TopologyEdge{
		{From: "ing-" + app.Name, To: "svc-" + app.Name, Kind: "route"},
		{From: "svc-" + app.Name, To: "deploy-" + app.Name, Kind: "select"},
		{From: "cfg-" + app.Name, To: "deploy-" + app.Name, Kind: "mount"},
	}
	return model.ApplicationTopology{
		AppID:        app.ID,
		AppName:      app.Name,
		Environment:  environment,
		Nodes:        nodes,
		Edges:        edges,
		LastSyncTime: time.Now(),
	}, nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func mergeVariables(base map[string]string, override map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		result[key] = strings.TrimSpace(v)
	}
	return result
}
