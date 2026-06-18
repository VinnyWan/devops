package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"devops-platform/internal/modules/log/model"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "log/repository"

// LogSourceAdapter abstracts a log backend
type LogSourceAdapter interface {
	Search(source *model.LogSource, req model.SearchRequest) (*model.SearchResponse, error)
	HealthCheck(source *model.LogSource) error
}

// LogRepo manages log sources and delegates to adapters
type LogRepo struct {
	db         *gorm.DB
	httpClient *http.Client
	adapters   map[string]LogSourceAdapter
}

func NewLogRepo(db *gorm.DB) *LogRepo {
	r := &LogRepo{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		adapters:   make(map[string]LogSourceAdapter),
	}
	r.adapters["elasticsearch"] = &ElasticsearchAdapter{httpClient: r.httpClient}
	return r
}

// --- Source CRUD ---

func (r *LogRepo) ListSources(page, pageSize int) ([]model.LogSource, int64, error) {
	var sources []model.LogSource
	var total int64
	q := r.db.Model(&model.LogSource{})
	q.Count(&total)
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&sources).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list log sources failed", err)
	}
	return sources, total, nil
}

func (r *LogRepo) GetSource(id uint) (*model.LogSource, error) {
	var src model.LogSource
	if err := r.db.First(&src, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get log source failed", err)
	}
	return &src, nil
}

func (r *LogRepo) SaveSource(src *model.LogSource) error {
	if err := r.db.Save(src).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save log source failed", err)
	}
	return nil
}

func (r *LogRepo) DeleteSource(id uint) error {
	if err := r.db.Delete(&model.LogSource{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete log source failed", err)
	}
	return nil
}

// --- Search (delegates to adapter) ---

func (r *LogRepo) Search(sourceID uint, req model.SearchRequest) (*model.SearchResponse, error) {
	src, err := r.GetSource(sourceID)
	if err != nil {
		return nil, obserr.Wrap("LOG_SOURCE_NOT_FOUND", op, "log source not found", err)
	}
	adapter, ok := r.adapters[src.Type]
	if !ok {
		return nil, obserr.New("LOG_UNSUPPORTED_TYPE", op, fmt.Sprintf("unsupported log source type: %s", src.Type))
	}
	return adapter.Search(src, req)
}

// --- Health check ---

func (r *LogRepo) TestConnection(id uint) error {
	src, err := r.GetSource(id)
	if err != nil {
		return obserr.Wrap("LOG_SOURCE_NOT_FOUND", op, "log source not found", err)
	}
	adapter, ok := r.adapters[src.Type]
	if !ok {
		return obserr.New("LOG_UNSUPPORTED_TYPE", op, fmt.Sprintf("unsupported type: %s", src.Type))
	}
	return adapter.HealthCheck(src)
}

// --- Elasticsearch Adapter ---

type ElasticsearchAdapter struct {
	httpClient *http.Client
}

type esSearchRequest struct {
	Size  int  `json:"size"`
	From  int  `json:"from"`
	Query struct {
		Bool struct {
			Must []interface{} `json:"must"`
		} `json:"bool"`
	} `json:"query"`
	Sort []map[string]interface{} `json:"sort"`
}

func (a *ElasticsearchAdapter) Search(source *model.LogSource, req model.SearchRequest) (*model.SearchResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	esReq := esSearchRequest{
		Size: req.PageSize,
		From: (req.Page - 1) * req.PageSize,
	}
	esReq.Sort = []map[string]interface{}{
		{"@timestamp": map[string]string{"order": "desc"}},
	}

	// Build query filters
	var must []interface{}

	// Time range
	if req.StartTime != "" || req.EndTime != "" {
		timeRange := map[string]interface{}{}
		if req.StartTime != "" {
			timeRange["gte"] = req.StartTime
		}
		if req.EndTime != "" {
			timeRange["lte"] = req.EndTime
		}
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{"@timestamp": timeRange},
		})
	}

	// Keywords (search across message field)
	if len(req.Keywords) > 0 {
		var should []interface{}
		for _, kw := range req.Keywords {
			should = append(should, map[string]interface{}{
				"match": map[string]string{"message": kw},
			})
			should = append(should, map[string]interface{}{
				"wildcard": map[string]string{"message": fmt.Sprintf("*%s*", kw)},
			})
		}
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{"should": should, "minimum_should_match": 1},
		})
	}

	// Level filter
	if req.Level != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]string{"level.keyword": strings.ToUpper(req.Level)},
		})
	}

	// Service filter
	if req.Service != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]string{"service.keyword": req.Service},
		})
	}

	if len(must) > 0 {
		esReq.Query.Bool.Must = must
	} else {
		// Match all if no filters
		esReq.Query.Bool.Must = []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}
	}

	body, err := json.Marshal(esReq)
	if err != nil {
		return nil, obserr.Wrap("LOG_SEARCH_FAILED", op, "failed to marshal ES query", err)
	}

	index := source.IndexPattern
	if index == "" {
		index = "app-logs-*"
	}
	u := strings.TrimRight(source.Endpoint, "/") + "/" + index + "/_search"

	httpReq, err := http.NewRequest("POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, obserr.Wrap("LOG_SEARCH_FAILED", op, "failed to build ES request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if source.Username != "" {
		httpReq.SetBasicAuth(source.Username, source.Password)
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, obserr.Wrap("LOG_SEARCH_FAILED", op, "ES search request failed", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, obserr.New("LOG_SEARCH_FAILED", op, fmt.Sprintf("ES returned %d: %s", resp.StatusCode, string(respBody)))
	}

	return a.parseESResponse(resp.Body, req.Page, req.PageSize)
}

func (a *ElasticsearchAdapter) HealthCheck(source *model.LogSource) error {
	u := strings.TrimRight(source.Endpoint, "/") + "/_cluster/health"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return obserr.New("LOG_CONNECT_FAILED", op, "failed to build health check request")
	}
	if source.Username != "" {
		req.SetBasicAuth(source.Username, source.Password)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return obserr.Wrap("LOG_CONNECT_FAILED", op, "ES health check failed", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return obserr.New("LOG_CONNECT_FAILED", op, fmt.Sprintf("ES returned %d", resp.StatusCode))
	}
	return nil
}

func (a *ElasticsearchAdapter) parseESResponse(body io.Reader, page, pageSize int) (*model.SearchResponse, error) {
	var raw struct {
		Hits struct {
			Total struct {
				Value    int64  `json:"value"`
				Relation string `json:"relation"`
			} `json:"total"`
			Hits []struct {
				Source struct {
					Timestamp string `json:"@timestamp"`
					Level     string `json:"level"`
					Service   string `json:"service"`
					Message   string `json:"message"`
					Host      string `json:"host"`
					TraceID   string `json:"traceId"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, obserr.Wrap("LOG_PARSE_FAILED", op, "failed to parse ES response", err)
	}

	var entries []model.LogEntry
	for _, h := range raw.Hits.Hits {
		entries = append(entries, model.LogEntry{
			Timestamp: h.Source.Timestamp,
			Level:     h.Source.Level,
			Service:   h.Source.Service,
			Message:   h.Source.Message,
			Host:      h.Source.Host,
			TraceID:   h.Source.TraceID,
		})
	}

	total := raw.Hits.Total.Value
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return &model.SearchResponse{
		Entries:    entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
