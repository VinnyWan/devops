package service

import (
	"devops-platform/internal/modules/log/model"
	"devops-platform/internal/modules/log/repository"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "log/service"

type LogService struct {
	repo *repository.LogRepo
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{repo: repository.NewLogRepo(db)}
}

func (s *LogService) ListSources(page, pageSize int) ([]model.LogSource, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListSources(page, pageSize)
}

func (s *LogService) GetSource(id uint) (*model.LogSource, error) {
	return s.repo.GetSource(id)
}

func (s *LogService) SaveSource(src *model.LogSource) error {
	if src.Endpoint == "" {
		return obserr.New("INVALID_PARAM", op, "endpoint is required")
	}
	if src.Name == "" {
		return obserr.New("INVALID_PARAM", op, "name is required")
	}
	if err := s.repo.TestConnection(src.ID); err != nil {
		src.Status = "error"
	} else {
		src.Status = "connected"
	}
	return s.repo.SaveSource(src)
}

func (s *LogService) DeleteSource(id uint) error {
	return s.repo.DeleteSource(id)
}

func (s *LogService) TestConnection(id uint) error {
	return s.repo.TestConnection(id)
}

func (s *LogService) Search(req model.SearchRequest) (*model.SearchResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	return s.repo.Search(req.SourceID, req)
}
