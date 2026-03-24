package service

import (
	"strings"

	"devops-platform/internal/modules/harbor/model"
	"devops-platform/internal/modules/harbor/repository"
	"devops-platform/internal/pkg/obserr"
	queryutil "devops-platform/internal/pkg/query"
)

type HarborService struct {
	repo *repository.HarborRepo
}

type ListProjectResponse struct {
	Total int                   `json:"total"`
	Items []model.HarborProject `json:"items"`
}

type ListImageResponse struct {
	Total int                     `json:"total"`
	Items []model.RepositoryImage `json:"items"`
}

type SaveHarborConfigRequest struct {
	Endpoint              string `json:"endpoint"`
	Project               string `json:"project"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	RobotToken            string `json:"robotToken"`
	TimeoutSeconds        int    `json:"timeoutSeconds"`
	TLSInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify"`
}

func NewHarborService() *HarborService {
	return &HarborService{repo: repository.NewHarborRepo()}
}

func (s *HarborService) ListProjects(keyword string) (ListProjectResponse, error) {
	if err := s.ValidateCurrentConfig(); err != nil {
		return ListProjectResponse{}, err
	}
	projects := s.repo.ListProjects()
	items := make([]model.HarborProject, 0, len(projects))
	for _, project := range projects {
		if !queryutil.MatchKeywordAny(keyword, project.Name) {
			continue
		}
		items = append(items, project)
	}
	return ListProjectResponse{Total: len(items), Items: items}, nil
}

func (s *HarborService) ListImages(projectName, keyword string) (ListImageResponse, error) {
	if err := s.ValidateCurrentConfig(); err != nil {
		return ListImageResponse{}, err
	}
	projectName = strings.TrimSpace(projectName)
	images := s.repo.ListImages(projectName)
	items := make([]model.RepositoryImage, 0, len(images))
	for _, image := range images {
		if !queryutil.MatchKeywordAny(keyword, image.Repository, image.Tag, image.Digest, image.ProjectName) {
			continue
		}
		items = append(items, image)
	}
	return ListImageResponse{Total: len(items), Items: items}, nil
}

func (s *HarborService) GetConfig() model.HarborConfig {
	return s.repo.GetConfig()
}

func (s *HarborService) SaveConfig(req SaveHarborConfigRequest) (model.HarborConfig, error) {
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		return model.HarborConfig{}, obserr.New("HARBOR_ENDPOINT_REQUIRED", "harbor.SaveConfig", "Harbor endpoint 不能为空")
	}
	project := strings.TrimSpace(req.Project)
	if project == "" {
		project = "library"
	}
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	config := model.HarborConfig{
		Endpoint:              endpoint,
		Project:               project,
		Username:              strings.TrimSpace(req.Username),
		Password:              req.Password,
		RobotToken:            strings.TrimSpace(req.RobotToken),
		TimeoutSeconds:        timeout,
		TLSInsecureSkipVerify: req.TLSInsecureSkipVerify,
	}
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return model.HarborConfig{}, obserr.Wrap("HARBOR_CONNECT_FAILED", "harbor.SaveConfig", "Harbor 配置连接失败", err)
	}
	saved := s.repo.SaveConfig(config)
	return saved, nil
}

func (s *HarborService) ValidateCurrentConfig() error {
	config := s.repo.GetConfig()
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return obserr.Wrap("HARBOR_CONNECT_FAILED", "harbor.ValidateCurrentConfig", "Harbor 配置连接失败", err)
	}
	return nil
}
