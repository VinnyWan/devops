package service

import (
	"strings"
	"time"

	"devops-platform/internal/modules/log/model"
	"devops-platform/internal/modules/log/repository"
	queryutil "devops-platform/internal/pkg/query"
)

type LogService struct {
	repo *repository.LogRepo
}

type SearchResponse struct {
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
	Items    []model.LogEntry `json:"items"`
}

func NewLogService() *LogService {
	return &LogService{repo: repository.NewLogRepo()}
}

func (s *LogService) Search(req model.SearchRequest) SearchResponse {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if !req.Start.IsZero() && !req.End.IsZero() && req.End.Before(req.Start) {
		req.Start, req.End = req.End, req.Start
	}

	keyword := req.Keyword
	source := strings.ToLower(strings.TrimSpace(req.Source))
	level := strings.ToLower(strings.TrimSpace(req.Level))
	items := make([]model.LogEntry, 0)
	for _, item := range s.repo.List() {
		if !queryutil.MatchKeywordAny(keyword, item.Message, item.Source, item.Level) {
			continue
		}
		if source != "" && strings.ToLower(item.Source) != source {
			continue
		}
		if level != "" && strings.ToLower(item.Level) != level {
			continue
		}
		if !req.Start.IsZero() && item.CreatedAt.Before(req.Start) {
			continue
		}
		if !req.End.IsZero() && item.CreatedAt.After(req.End) {
			continue
		}
		items = append(items, item)
	}

	total := len(items)
	startIndex := (req.Page - 1) * req.PageSize
	if startIndex >= total {
		return SearchResponse{
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
			Items:    []model.LogEntry{},
		}
	}
	endIndex := startIndex + req.PageSize
	if endIndex > total {
		endIndex = total
	}

	return SearchResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Items:    items[startIndex:endIndex],
	}
}

func ParseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
