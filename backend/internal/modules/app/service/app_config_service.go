package service

import (
	"errors"
	"regexp"

	"devops-platform/internal/modules/app/model"
	"devops-platform/internal/modules/app/repository"
)

type AppConfigService struct {
	repo *repository.AppRepo
}

func NewAppConfigService() *AppConfigService {
	return NewAppConfigServiceWithRepo(repository.NewAppRepo())
}

func NewAppConfigServiceWithRepo(repo *repository.AppRepo) *AppConfigService {
	if repo == nil {
		repo = repository.NewAppRepo()
	}
	return &AppConfigService{repo: repo}
}

func (s *AppConfigService) GetAppConfig(appID uint) (model.AppConfig, error) {
	config, ok := s.repo.GetAppConfig(appID)
	if !ok {
		return model.AppConfig{}, errors.New("应用配置不存在")
	}
	return config, nil
}

func (s *AppConfigService) SaveAppConfig(config model.AppConfig) (model.AppConfig, error) {
	return s.repo.SaveAppConfig(config), nil
}

func (s *AppConfigService) DeleteAppConfig(appID uint) bool {
	return s.repo.DeleteAppConfig(appID)
}

func (s *AppConfigService) GetBuildConfig(appID uint) (model.BuildConfig, error) {
	config, ok := s.repo.GetBuildConfig(appID)
	if !ok {
		return model.BuildConfig{}, errors.New("构建配置不存在")
	}
	return config, nil
}

func (s *AppConfigService) SaveBuildConfig(config model.BuildConfig) (model.BuildConfig, error) {
	return s.repo.SaveBuildConfig(config), nil
}

func (s *AppConfigService) DeleteBuildConfig(appID uint) bool {
	return s.repo.DeleteBuildConfig(appID)
}

func (s *AppConfigService) GetDeployConfig(appID uint, environment string) (model.DeployConfig, error) {
	config, ok := s.repo.GetDeployConfig(appID, environment)
	if !ok {
		return model.DeployConfig{}, errors.New("部署配置不存在")
	}
	return config, nil
}

// GetDeployConfigByAppID 获取应用的部署配置（兼容旧接口）
func (s *AppConfigService) GetDeployConfigByAppID(appID uint) (model.DeployConfig, error) {
	config, ok := s.repo.GetDeployConfigByAppID(appID)
	if !ok {
		return model.DeployConfig{}, errors.New("部署配置不存在")
	}
	return config, nil
}

// ListDeployConfigsByApp 获取应用所有环境的部署配置
func (s *AppConfigService) ListDeployConfigsByApp(appID uint) []model.DeployConfig {
	return s.repo.ListDeployConfigsByApp(appID)
}

func (s *AppConfigService) SaveDeployConfig(config model.DeployConfig) (model.DeployConfig, error) {
	return s.repo.SaveDeployConfig(config), nil
}

// DeleteDeployConfigByEnv 删除指定环境的部署配置
func (s *AppConfigService) DeleteDeployConfigByEnv(appID uint, environment string) bool {
	return s.repo.DeleteDeployConfigByEnv(appID, environment)
}

func (s *AppConfigService) DeleteDeployConfig(appID uint) bool {
	return s.repo.DeleteDeployConfig(appID)
}

func (s *AppConfigService) GetTechStackConfig(appID uint) (model.TechStackConfig, error) {
	config, ok := s.repo.GetTechStackConfig(appID)
	if !ok {
		return model.TechStackConfig{}, errors.New("技术栈配置不存在")
	}
	return config, nil
}

func (s *AppConfigService) SaveTechStackConfig(config model.TechStackConfig) (model.TechStackConfig, error) {
	return s.repo.SaveTechStackConfig(config), nil
}

func (s *AppConfigService) DeleteTechStackConfig(appID uint) bool {
	return s.repo.DeleteTechStackConfig(appID)
}

// ========== 新增方法 ==========

// ListAppsWithFilter 分页查询应用列表，支持搜索和筛选
func (s *AppConfigService) ListAppsWithFilter(page, pageSize int, keyword, instanceType, status string) ([]model.AppConfig, int64) {
	return s.repo.ListAppsWithFilter(page, pageSize, keyword, instanceType, status)
}

// ToggleAppStatus 切换应用状态
func (s *AppConfigService) ToggleAppStatus(appID uint, status string) (model.AppConfig, error) {
	// 状态值校验
	if status != model.StatusRunning && status != model.StatusOffline {
		return model.AppConfig{}, errors.New("无效的状态值，只支持 running 或 offline")
	}
	config, ok := s.repo.ToggleAppStatus(appID, status)
	if !ok {
		return model.AppConfig{}, errors.New("应用配置不存在")
	}
	return config, nil
}

// GetContainerConfig 获取指定环境的容器配置
func (s *AppConfigService) GetContainerConfig(appID uint, environment string) (model.ContainerConfig, error) {
	config, ok := s.repo.GetContainerConfig(appID, environment)
	if !ok {
		return model.ContainerConfig{}, errors.New("容器配置不存在")
	}
	return config, nil
}

// GetContainerConfigByAppID 获取应用的容器配置（兼容旧接口）
func (s *AppConfigService) GetContainerConfigByAppID(appID uint) (model.ContainerConfig, error) {
	config, ok := s.repo.GetContainerConfigByAppID(appID)
	if !ok {
		return model.ContainerConfig{}, errors.New("容器配置不存在")
	}
	return config, nil
}

// ListContainerConfigsByApp 获取应用所有环境的容器配置
func (s *AppConfigService) ListContainerConfigsByApp(appID uint) []model.ContainerConfig {
	return s.repo.ListContainerConfigsByApp(appID)
}

// SaveContainerConfig 保存容器配置
func (s *AppConfigService) SaveContainerConfig(config model.ContainerConfig) (model.ContainerConfig, error) {
	return s.repo.SaveContainerConfig(config), nil
}

// DeleteContainerConfigByEnv 删除指定环境的容器配置
func (s *AppConfigService) DeleteContainerConfigByEnv(appID uint, environment string) bool {
	return s.repo.DeleteContainerConfigByEnv(appID, environment)
}

// DeleteContainerConfig 删除容器配置
func (s *AppConfigService) DeleteContainerConfig(appID uint) bool {
	return s.repo.DeleteContainerConfig(appID)
}

// ValidateAppConfig 验证应用配置
func ValidateAppConfig(config model.AppConfig) error {
	if config.Name == "" {
		return errors.New("应用名称不能为空")
	}
	if config.Domain != "" && !isValidDomain(config.Domain) {
		return errors.New("域名格式错误")
	}
	return nil
}

// isValidDomain 验证域名格式
func isValidDomain(domain string) bool {
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	matched, _ := regexp.MatchString(pattern, domain)
	return matched
}

// ValidateBuildConfig 验证构建配置
func ValidateBuildConfig(config model.BuildConfig) error {
	if config.Dockerfile == "" {
		return errors.New("Dockerfile不能为空")
	}
	return nil
}

// ValidateDeployConfig 验证部署配置
func ValidateDeployConfig(config model.DeployConfig) error {
	if config.ServicePort <= 0 || config.ServicePort > 65535 {
		return errors.New("服务端口必须在1-65535范围内")
	}
	if config.CPURequest == "" || config.CPULimit == "" {
		return errors.New("CPU配置无效")
	}
	return nil
}

// ValidateTechStackConfig 验证技术栈配置
func ValidateTechStackConfig(config model.TechStackConfig) error {
	if config.Name == "" {
		return errors.New("技术栈名称不能为空")
	}
	if config.Language == "" {
		return errors.New("编程语言不能为空")
	}
	return nil
}

// ValidateContainerConfig 验证容器配置
func ValidateContainerConfig(config model.ContainerConfig) error {
	if config.Image == "" {
		return errors.New("镜像地址不能为空")
	}
	if config.CPURequest == "" || config.CPULimit == "" {
		return errors.New("CPU配置不能为空")
	}
	if config.MemoryRequest == "" || config.MemoryLimit == "" {
		return errors.New("内存配置不能为空")
	}
	return nil
}

// GetDefaultTechStack 获取默认技术栈配置
func GetDefaultTechStack(stackType string) model.TechStackConfig {
	return model.TechStackConfig{
		Language: stackType,
		Version:  "",
	}
}

// GetDefaultBuildConfig 获取默认构建配置
func GetDefaultBuildConfig(buildType string) model.BuildConfig {
	return model.BuildConfig{
		BuildTool: buildType,
	}
}

// GetDefaultDeployConfig 获取默认部署配置
func GetDefaultDeployConfig() model.DeployConfig {
	return model.DeployConfig{}
}
