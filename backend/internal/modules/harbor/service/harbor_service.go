package service

import (
	"devops-platform/internal/modules/harbor/model"
	"devops-platform/internal/modules/harbor/repository"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "harbor/service"

type HarborService struct {
	repo *repository.HarborRepo
}

func NewHarborService(db *gorm.DB) *HarborService {
	return &HarborService{repo: repository.NewHarborRepo(db)}
}

func (s *HarborService) ListConfigs(page, pageSize int) ([]model.HarborConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListConfigs(page, pageSize)
}

func (s *HarborService) SaveConfig(cfg *model.HarborConfig) error {
	if cfg.URL == "" {
		return obserr.New("INVALID_PARAM", op, "url is required")
	}
	if cfg.Username == "" {
		return obserr.New("INVALID_PARAM", op, "username is required")
	}
	if cfg.Password == "" {
		return obserr.New("INVALID_PARAM", op, "password is required")
	}
	if err := s.repo.TestConnection(cfg.URL, cfg.Username, cfg.Password); err != nil {
		cfg.Status = "error"
	} else {
		cfg.Status = "connected"
	}
	return s.repo.SaveConfig(cfg)
}

func (s *HarborService) DeleteConfig(id uint) error {
	return s.repo.DeleteConfig(id)
}

func (s *HarborService) TestConnection(url, username, password string) error {
	if url == "" {
		return obserr.New("INVALID_PARAM", op, "url is required")
	}
	return s.repo.TestConnection(url, username, password)
}

func (s *HarborService) ListProjects(configID uint, keyword string, page, pageSize int) ([]model.Project, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListProjects(configID, keyword, page, pageSize)
}

func (s *HarborService) ListRepositories(configID uint, projectName, keyword string, page, pageSize int) ([]model.Repository, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListRepositories(configID, projectName, keyword, page, pageSize)
}

func (s *HarborService) ListArtifacts(configID uint, projectName, repoName string, page, pageSize int) ([]model.Artifact, int64, error) {
	return s.repo.ListArtifacts(configID, projectName, repoName, page, pageSize)
}

func (s *HarborService) DeleteArtifact(configID uint, projectName, repoName, reference string) error {
	if reference == "" {
		return obserr.New("INVALID_PARAM", op, "reference (tag or digest) is required")
	}
	return s.repo.DeleteArtifact(configID, projectName, repoName, reference)
}
