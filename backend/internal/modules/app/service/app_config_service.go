package service

import (
	"errors"

	"devops-platform/internal/modules/app/model"
	"devops-platform/internal/modules/app/repository"
)

type AppConfigService struct {
	repo *repository.AppRepo
}

func NewAppConfigService() *AppConfigService {
	return &AppConfigService{repo: repository.NewAppRepo()}
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

func (s *AppConfigService) GetDeployConfig(appID uint) (model.DeployConfig, error) {
	config, ok := s.repo.GetDeployConfig(appID)
	if !ok {
		return model.DeployConfig{}, errors.New("部署配置不存在")
	}
	return config, nil
}

func (s *AppConfigService) SaveDeployConfig(config model.DeployConfig) (model.DeployConfig, error) {
	return s.repo.SaveDeployConfig(config), nil
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

// ValidateAppConfig 验证应用配置
func ValidateAppConfig(config model.AppConfig) error {
	if config.Name == "" {
		return errors.New("应用名称不能为空")
	}
	return nil
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
		return errors.New("端口号必须在1-65535之间")
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
