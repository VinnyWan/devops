package service

import (
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
)

type SnippetService struct {
	repo *repository.SnippetRepo
}

func NewSnippetService(repo *repository.SnippetRepo) *SnippetService {
	return &SnippetService{repo: repo}
}

type SnippetListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

func (s *SnippetService) List(tenantID, userID uint, page, pageSize int) ([]model.CommandSnippet, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.ListInTenant(tenantID, userID, page, pageSize)
}

func (s *SnippetService) GetByID(id uint) (*model.CommandSnippet, error) {
	return s.repo.GetByID(id)
}

type SnippetCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Tags       string `json:"tags"`
	Visibility string `json:"visibility"`
}

func (s *SnippetService) Create(tenantID, userID uint, req SnippetCreateRequest) (*model.CommandSnippet, error) {
	visibility := req.Visibility
	if visibility == "" {
		visibility = "personal"
	}
	snippet := &model.CommandSnippet{
		TenantID:   tenantID,
		UserID:     userID,
		Name:       req.Name,
		Content:    req.Content,
		Tags:       req.Tags,
		Visibility: visibility,
	}
	if err := s.repo.Create(snippet); err != nil {
		return nil, err
	}
	return snippet, nil
}

type SnippetUpdateRequest struct {
	ID         uint   `json:"id" binding:"required"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Tags       string `json:"tags"`
	Visibility string `json:"visibility"`
}

func (s *SnippetService) Update(tenantID, userID uint, req SnippetUpdateRequest) (*model.CommandSnippet, error) {
	snippet, err := s.repo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}
	if snippet.TenantID != tenantID || (snippet.UserID != userID && snippet.Visibility == "personal") {
		return nil, ErrPermissionDenied
	}
	if req.Name != "" {
		snippet.Name = req.Name
	}
	if req.Content != "" {
		snippet.Content = req.Content
	}
	snippet.Tags = req.Tags
	if req.Visibility != "" {
		snippet.Visibility = req.Visibility
	}
	if err := s.repo.Update(snippet); err != nil {
		return nil, err
	}
	return snippet, nil
}

func (s *SnippetService) Delete(tenantID, userID uint, id uint) error {
	snippet, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if snippet.TenantID != tenantID || (snippet.UserID != userID && snippet.Visibility == "personal") {
		return ErrPermissionDenied
	}
	return s.repo.Delete(id)
}

func (s *SnippetService) Search(tenantID, userID uint, keyword string) ([]model.CommandSnippet, error) {
	return s.repo.Search(tenantID, userID, keyword, 20)
}

var ErrPermissionDenied = &permissionError{}

type permissionError struct{}

func (e *permissionError) Error() string { return "权限不足" }
