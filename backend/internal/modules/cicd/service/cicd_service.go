package service

import (
	"devops-platform/internal/modules/cicd/model"
	"devops-platform/internal/modules/cicd/repository"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "cicd/service"

type CICDService struct {
	repo *repository.CICDRepo
	db   *gorm.DB
}

func NewCICDService(db *gorm.DB) *CICDService {
	return &CICDService{repo: repository.NewCICDRepo(db), db: db}
}

// Config management
func (s *CICDService) ListConfigs(page, pageSize int) ([]model.JenkinsConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListConfigs(page, pageSize)
}

func (s *CICDService) SaveConfig(cfg *model.JenkinsConfig) error {
	if cfg.URL == "" {
		return obserr.New("INVALID_PARAM", op, "jenkins url is required")
	}
	if cfg.Username == "" {
		return obserr.New("INVALID_PARAM", op, "username is required")
	}
	if cfg.APIToken == "" {
		return obserr.New("INVALID_PARAM", op, "api token is required")
	}
	if err := s.repo.TestConnection(cfg.URL, cfg.Username, cfg.APIToken); err != nil {
		cfg.Status = "error"
	} else {
		cfg.Status = "connected"
	}
	return s.repo.SaveConfig(cfg)
}

func (s *CICDService) DeleteConfig(id uint) error {
	return s.repo.DeleteConfig(id)
}

func (s *CICDService) TestConnection(url, username, apiToken string) error {
	if url == "" {
		return obserr.New("INVALID_PARAM", op, "url is required")
	}
	return s.repo.TestConnection(url, username, apiToken)
}

// Job management
func (s *CICDService) ListJobs(configID uint, keyword string) ([]model.JobInfo, error) {
	return s.repo.ListJobs(configID, keyword)
}

func (s *CICDService) TriggerBuild(configID uint, jobName string) error {
	if jobName == "" {
		return obserr.New("INVALID_PARAM", op, "job name is required")
	}
	return s.repo.TriggerBuild(configID, jobName)
}

func (s *CICDService) ListBuilds(configID uint, jobName string) ([]model.BuildInfo, error) {
	return s.repo.ListBuilds(configID, jobName)
}

func (s *CICDService) GetBuildLog(configID uint, jobName string, buildNumber int) (*model.BuildLogEntry, error) {
	return s.repo.GetBuildLog(configID, jobName, buildNumber)
}

// Pipeline CRUD
func (s *CICDService) ListPipelines(page, pageSize int) ([]model.Pipeline, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListPipelines(page, pageSize)
}

func (s *CICDService) SavePipeline(p *model.Pipeline) error {
	if p.Name == "" {
		return obserr.New("INVALID_PARAM", op, "pipeline name is required")
	}
	return s.repo.SavePipeline(p)
}

func (s *CICDService) DeletePipeline(id uint) error {
	return s.repo.DeletePipeline(id)
}
