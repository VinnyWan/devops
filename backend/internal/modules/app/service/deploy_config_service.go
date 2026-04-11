package service

type DeployConfigService struct{}

func NewDeployConfigService() *DeployConfigService {
	return &DeployConfigService{}
}

func (s *DeployConfigService) ListK8sClusters() []map[string]interface{} {
	return []map[string]interface{}{}
}

func (s *DeployConfigService) ListImageVersions(appName string) []map[string]interface{} {
	return []map[string]interface{}{}
}
