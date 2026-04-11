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
	return NewAppServiceWithRepo(repository.NewAppRepo())
}

func NewAppServiceWithRepo(repo *repository.AppRepo) *AppService {
	if repo == nil {
		repo = repository.NewAppRepo()
	}
	return &AppService{repo: repo}
}

func (s *AppService) List() []model.Application {
	return s.ListInTenant(0)
}

func (s *AppService) ListInTenant(tenantID uint) []model.Application {
	return s.repo.ListInTenant(tenantID)
}

func (s *AppService) ListTemplates(keyword string) ListTemplatesResponse {
	return s.ListTemplatesInTenant(0, keyword)
}

func (s *AppService) ListTemplatesInTenant(tenantID uint, keyword string) ListTemplatesResponse {
	templates := s.repo.ListTemplatesInTenant(tenantID)
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
	return s.SaveTemplateInTenant(0, req)
}

func (s *AppService) SaveTemplateInTenant(tenantID uint, req SaveTemplateRequest) (model.AppTemplate, error) {
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
	template := s.repo.SaveTemplateInTenant(tenantID, model.AppTemplate{
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
	return s.DeployInTenant(0, req)
}

func (s *AppService) DeployInTenant(tenantID uint, req DeployRequest) (model.ApplicationDeployment, error) {
	app, ok := s.repo.FindAppByIDInTenant(tenantID, req.AppID)
	if !ok {
		return model.ApplicationDeployment{}, errors.New("应用不存在")
	}
	template, ok := s.repo.FindTemplateByIDInTenant(tenantID, req.TemplateID)
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
	deployment := s.repo.CreateDeploymentInTenant(tenantID, model.ApplicationDeployment{
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
	s.repo.CreateVersionInTenant(tenantID, model.ApplicationVersion{
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
	return s.ListDeploymentsInTenant(0, appID, environment, limit)
}

func (s *AppService) ListDeploymentsInTenant(tenantID uint, appID uint, environment string, limit int) ListDeploymentsResponse {
	if limit <= 0 {
		limit = 20
	}
	items := s.repo.ListDeploymentsInTenant(tenantID, appID, environment)
	if len(items) > limit {
		items = items[:limit]
	}
	return ListDeploymentsResponse{Total: len(items), Items: items}
}

func (s *AppService) ListVersions(appID uint, limit int) ListVersionsResponse {
	return s.ListVersionsInTenant(0, appID, limit)
}

func (s *AppService) ListVersionsInTenant(tenantID uint, appID uint, limit int) ListVersionsResponse {
	if limit <= 0 {
		limit = 20
	}
	items := s.repo.ListVersionsInTenant(tenantID, appID)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return ListVersionsResponse{Total: len(items), Items: items}
}

func (s *AppService) Rollback(req RollbackRequest) (model.ApplicationVersion, error) {
	return s.RollbackInTenant(0, req)
}

func (s *AppService) RollbackInTenant(tenantID uint, req RollbackRequest) (model.ApplicationVersion, error) {
	app, ok := s.repo.FindAppByIDInTenant(tenantID, req.AppID)
	if !ok {
		return model.ApplicationVersion{}, errors.New("应用不存在")
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		return model.ApplicationVersion{}, errors.New("目标版本不能为空")
	}
	version, ok := s.repo.FindVersionInTenant(tenantID, app.ID, target)
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
	rolled := s.repo.CreateVersionInTenant(tenantID, model.ApplicationVersion{
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
	return s.QueryTopologyInTenant(0, appID, environment)
}

func (s *AppService) QueryTopologyInTenant(tenantID uint, appID uint, environment string) (model.ApplicationTopology, error) {
	app, ok := s.repo.FindAppByIDInTenant(tenantID, appID)
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

// ========== 应用配置相关方法 ==========

func (s *AppService) GetAppConfig(appID uint) (model.AppConfig, error) {
	return s.GetAppConfigInTenant(0, appID)
}

func (s *AppService) GetAppConfigInTenant(tenantID uint, appID uint) (model.AppConfig, error) {
	config, ok := s.repo.GetAppConfigInTenant(tenantID, appID)
	if !ok {
		return model.AppConfig{}, errors.New("应用配置不存在")
	}
	return config, nil
}

func (s *AppService) SaveAppConfig(config model.AppConfig) (model.AppConfig, error) {
	return s.SaveAppConfigInTenant(0, config)
}

func (s *AppService) SaveAppConfigInTenant(tenantID uint, config model.AppConfig) (model.AppConfig, error) {
	return s.repo.SaveAppConfigInTenant(tenantID, config), nil
}

func (s *AppService) GetBuildConfig(appID uint) (model.BuildConfig, error) {
	return s.GetBuildConfigInTenant(0, appID)
}

func (s *AppService) GetBuildConfigInTenant(tenantID uint, appID uint) (model.BuildConfig, error) {
	config, ok := s.repo.GetBuildConfigInTenant(tenantID, appID)
	if !ok {
		return model.BuildConfig{}, errors.New("构建配置不存在")
	}
	return config, nil
}

func (s *AppService) SaveBuildConfig(config model.BuildConfig) (model.BuildConfig, error) {
	return s.SaveBuildConfigInTenant(0, config)
}

func (s *AppService) SaveBuildConfigInTenant(tenantID uint, config model.BuildConfig) (model.BuildConfig, error) {
	return s.repo.SaveBuildConfigInTenant(tenantID, config), nil
}

// GetDeployConfig 获取指定环境的部署配置
func (s *AppService) GetDeployConfig(appID uint, environment string) (model.DeployConfig, error) {
	return s.GetDeployConfigInTenant(0, appID, environment)
}

func (s *AppService) GetDeployConfigInTenant(tenantID uint, appID uint, environment string) (model.DeployConfig, error) {
	config, ok := s.repo.GetDeployConfigInTenant(tenantID, appID, environment)
	if !ok {
		return model.DeployConfig{}, errors.New("部署配置不存在")
	}
	return config, nil
}

// GetDeployConfigByAppID 获取应用的部署配置（兼容旧接口）
func (s *AppService) GetDeployConfigByAppID(appID uint) (model.DeployConfig, error) {
	return s.GetDeployConfigByAppIDInTenant(0, appID)
}

func (s *AppService) GetDeployConfigByAppIDInTenant(tenantID uint, appID uint) (model.DeployConfig, error) {
	config, ok := s.repo.GetDeployConfigByAppIDInTenant(tenantID, appID)
	if !ok {
		return model.DeployConfig{}, errors.New("部署配置不存在")
	}
	return config, nil
}

// ListDeployConfigsByApp 获取应用所有环境的部署配置
func (s *AppService) ListDeployConfigsByApp(appID uint) []model.DeployConfig {
	return s.ListDeployConfigsByAppInTenant(0, appID)
}

func (s *AppService) ListDeployConfigsByAppInTenant(tenantID uint, appID uint) []model.DeployConfig {
	return s.repo.ListDeployConfigsByAppInTenant(tenantID, appID)
}

// SaveDeployConfig 保存部署配置
func (s *AppService) SaveDeployConfig(config model.DeployConfig) (model.DeployConfig, error) {
	return s.SaveDeployConfigInTenant(0, config)
}

func (s *AppService) SaveDeployConfigInTenant(tenantID uint, config model.DeployConfig) (model.DeployConfig, error) {
	return s.repo.SaveDeployConfigInTenant(tenantID, config), nil
}

// DeleteDeployConfigByEnv 删除指定环境的部署配置
func (s *AppService) DeleteDeployConfigByEnv(appID uint, environment string) bool {
	return s.DeleteDeployConfigByEnvInTenant(0, appID, environment)
}

func (s *AppService) DeleteDeployConfigByEnvInTenant(tenantID uint, appID uint, environment string) bool {
	return s.repo.DeleteDeployConfigByEnvInTenant(tenantID, appID, environment)
}

func (s *AppService) GetTechStackConfig(appID uint) (model.TechStackConfig, error) {
	return s.GetTechStackConfigInTenant(0, appID)
}

func (s *AppService) GetTechStackConfigInTenant(tenantID uint, appID uint) (model.TechStackConfig, error) {
	config, ok := s.repo.GetTechStackConfigInTenant(tenantID, appID)
	if !ok {
		return model.TechStackConfig{}, errors.New("技术栈配置不存在")
	}
	return config, nil
}

func (s *AppService) SaveTechStackConfig(config model.TechStackConfig) (model.TechStackConfig, error) {
	return s.SaveTechStackConfigInTenant(0, config)
}

func (s *AppService) SaveTechStackConfigInTenant(tenantID uint, config model.TechStackConfig) (model.TechStackConfig, error) {
	return s.repo.SaveTechStackConfigInTenant(tenantID, config), nil
}

// ========== 删除配置相关方法 ==========

func (s *AppService) DeleteAppConfig(appID uint) bool {
	return s.DeleteAppConfigInTenant(0, appID)
}

func (s *AppService) DeleteAppConfigInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteAppConfigInTenant(tenantID, appID)
}

// DeleteAppConfigCascade 级联删除应用及其所有关联配置
func (s *AppService) DeleteAppConfigCascade(appID uint) bool {
	return s.DeleteAppConfigCascadeInTenant(0, appID)
}

func (s *AppService) DeleteAppConfigCascadeInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteAppConfigCascadeInTenant(tenantID, appID)
}

func (s *AppService) DeleteBuildConfig(appID uint) bool {
	return s.DeleteBuildConfigInTenant(0, appID)
}

func (s *AppService) DeleteBuildConfigInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteBuildConfigInTenant(tenantID, appID)
}

func (s *AppService) DeleteDeployConfig(appID uint) bool {
	return s.DeleteDeployConfigInTenant(0, appID)
}

func (s *AppService) DeleteDeployConfigInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteDeployConfigInTenant(tenantID, appID)
}

func (s *AppService) DeleteTechStackConfig(appID uint) bool {
	return s.DeleteTechStackConfigInTenant(0, appID)
}

func (s *AppService) DeleteTechStackConfigInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteTechStackConfigInTenant(tenantID, appID)
}

// ========== 新增方法 ==========

// ListAppsWithFilter 分页查询应用列表，支持搜索和筛选
func (s *AppService) ListAppsWithFilter(page, pageSize int, keyword, instanceType, status string) ([]model.AppConfig, int64) {
	return s.ListAppsWithFilterInTenant(0, page, pageSize, keyword, instanceType, status)
}

func (s *AppService) ListAppsWithFilterInTenant(tenantID uint, page, pageSize int, keyword, instanceType, status string) ([]model.AppConfig, int64) {
	return s.repo.ListAppsWithFilterInTenant(tenantID, page, pageSize, keyword, instanceType, status)
}

// ToggleAppStatus 切换应用状态
func (s *AppService) ToggleAppStatus(appID uint, status string) (model.AppConfig, error) {
	return s.ToggleAppStatusInTenant(0, appID, status)
}

func (s *AppService) ToggleAppStatusInTenant(tenantID uint, appID uint, status string) (model.AppConfig, error) {
	// 状态值校验
	if status != model.StatusRunning && status != model.StatusOffline {
		return model.AppConfig{}, errors.New("无效的状态值，只支持 running 或 offline")
	}
	config, ok := s.repo.ToggleAppStatusInTenant(tenantID, appID, status)
	if !ok {
		return model.AppConfig{}, errors.New("应用配置不存在")
	}
	return config, nil
}

// GetContainerConfig 获取指定环境的容器配置
func (s *AppService) GetContainerConfig(appID uint, environment string) (model.ContainerConfig, error) {
	return s.GetContainerConfigInTenant(0, appID, environment)
}

func (s *AppService) GetContainerConfigInTenant(tenantID uint, appID uint, environment string) (model.ContainerConfig, error) {
	config, ok := s.repo.GetContainerConfigInTenant(tenantID, appID, environment)
	if !ok {
		return model.ContainerConfig{}, errors.New("容器配置不存在")
	}
	return config, nil
}

// GetContainerConfigByAppID 获取应用的容器配置（兼容旧接口）
func (s *AppService) GetContainerConfigByAppID(appID uint) (model.ContainerConfig, error) {
	return s.GetContainerConfigByAppIDInTenant(0, appID)
}

func (s *AppService) GetContainerConfigByAppIDInTenant(tenantID uint, appID uint) (model.ContainerConfig, error) {
	config, ok := s.repo.GetContainerConfigByAppIDInTenant(tenantID, appID)
	if !ok {
		return model.ContainerConfig{}, errors.New("容器配置不存在")
	}
	return config, nil
}

// ListContainerConfigsByApp 获取应用所有环境的容器配置
func (s *AppService) ListContainerConfigsByApp(appID uint) []model.ContainerConfig {
	return s.ListContainerConfigsByAppInTenant(0, appID)
}

func (s *AppService) ListContainerConfigsByAppInTenant(tenantID uint, appID uint) []model.ContainerConfig {
	return s.repo.ListContainerConfigsByAppInTenant(tenantID, appID)
}

// SaveContainerConfig 保存容器配置
func (s *AppService) SaveContainerConfig(config model.ContainerConfig) (model.ContainerConfig, error) {
	return s.SaveContainerConfigInTenant(0, config)
}

func (s *AppService) SaveContainerConfigInTenant(tenantID uint, config model.ContainerConfig) (model.ContainerConfig, error) {
	return s.repo.SaveContainerConfigInTenant(tenantID, config), nil
}

// DeleteContainerConfigByEnv 删除指定环境的容器配置
func (s *AppService) DeleteContainerConfigByEnv(appID uint, environment string) bool {
	return s.DeleteContainerConfigByEnvInTenant(0, appID, environment)
}

func (s *AppService) DeleteContainerConfigByEnvInTenant(tenantID uint, appID uint, environment string) bool {
	return s.repo.DeleteContainerConfigByEnvInTenant(tenantID, appID, environment)
}

// DeleteContainerConfig 删除容器配置
func (s *AppService) DeleteContainerConfig(appID uint) bool {
	return s.DeleteContainerConfigInTenant(0, appID)
}

func (s *AppService) DeleteContainerConfigInTenant(tenantID uint, appID uint) bool {
	return s.repo.DeleteContainerConfigInTenant(tenantID, appID)
}
