package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "monitor/repository"

// MonitorRepo handles DB access and Prometheus HTTP communication
type MonitorRepo struct {
	db         *gorm.DB
	httpClient *http.Client
}

// NewMonitorRepo creates a new MonitorRepo with the given DB and a default HTTP client
func NewMonitorRepo(db *gorm.DB) *MonitorRepo {
	return &MonitorRepo{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Prometheus Config CRUD ---

// ListConfigs returns a paginated list of PrometheusConfig records
func (r *MonitorRepo) ListConfigs(page, pageSize int) ([]model.PrometheusConfig, int64, error) {
	var configs []model.PrometheusConfig
	var total int64
	q := r.db.Model(&model.PrometheusConfig{})
	q.Count(&total)
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, 0, obserr.Wrap("DB_ERROR", op, "list prometheus configs failed", err)
	}
	return configs, total, nil
}

// GetConfig retrieves a single PrometheusConfig by ID
func (r *MonitorRepo) GetConfig(id uint) (*model.PrometheusConfig, error) {
	var cfg model.PrometheusConfig
	if err := r.db.First(&cfg, id).Error; err != nil {
		return nil, obserr.Wrap("DB_ERROR", op, "get prometheus config failed", err)
	}
	return &cfg, nil
}

// SaveConfig creates or updates a PrometheusConfig record
func (r *MonitorRepo) SaveConfig(cfg *model.PrometheusConfig) error {
	if err := r.db.Save(cfg).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "save prometheus config failed", err)
	}
	return nil
}

// DeleteConfig soft-deletes a PrometheusConfig by ID
func (r *MonitorRepo) DeleteConfig(id uint) error {
	if err := r.db.Delete(&model.PrometheusConfig{}, id).Error; err != nil {
		return obserr.Wrap("DB_ERROR", op, "delete prometheus config failed", err)
	}
	return nil
}

// --- Connection Test ---

// TestConnection tries to reach a Prometheus endpoint's /api/v1/status/buildinfo
func (r *MonitorRepo) TestConnection(endpoint, username, password string) error {
	req, err := http.NewRequest("GET", endpoint+"/api/v1/status/buildinfo", nil)
	if err != nil {
		return obserr.New("PROMETHEUS_CONNECT_FAILED", op, "failed to build request to prometheus")
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return obserr.Wrap("PROMETHEUS_CONNECT_FAILED", op, "cannot reach prometheus endpoint", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return obserr.New("PROMETHEUS_CONNECT_FAILED", op, fmt.Sprintf("prometheus returned %d: %s", resp.StatusCode, string(body)))
	}
	return nil
}

// --- PromQL Query ---

// QueryInstant performs an instant PromQL query against the config's Prometheus
func (r *MonitorRepo) QueryInstant(configID uint, promQL string) (*model.MetricQueryResponse, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, obserr.Wrap("PROMETHEUS_CONFIG_NOT_FOUND", op, "config not found", err)
	}
	return r.executeQuery(cfg, promQL, "query", "")
}

// QueryRange performs a range PromQL query against the config's Prometheus
func (r *MonitorRepo) QueryRange(configID uint, promQL, start, end, step string) (*model.MetricQueryResponse, error) {
	cfg, err := r.GetConfig(configID)
	if err != nil {
		return nil, obserr.Wrap("PROMETHEUS_CONFIG_NOT_FOUND", op, "config not found", err)
	}
	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	params.Set("step", step)
	return r.executeQuery(cfg, promQL, "query_range", params.Encode())
}

func (r *MonitorRepo) executeQuery(cfg *model.PrometheusConfig, promQL, promEndpoint, params string) (*model.MetricQueryResponse, error) {
	u := fmt.Sprintf("%s/api/v1/%s?query=%s", cfg.Endpoint, promEndpoint, url.QueryEscape(promQL))
	if params != "" {
		u += "&" + params
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, obserr.Wrap("PROMETHEUS_QUERY_FAILED", op, "failed to build prometheus query", err)
	}
	if cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, obserr.Wrap("PROMETHEUS_QUERY_FAILED", op, "prometheus query failed", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, obserr.New("PROMETHEUS_QUERY_FAILED", op, fmt.Sprintf("prometheus returned %d: %s", resp.StatusCode, string(body)))
	}
	return r.parsePrometheusResponse(resp.Body)
}

// parsePrometheusResponse parses the Prometheus HTTP API JSON response
func (r *MonitorRepo) parsePrometheusResponse(body io.Reader) (*model.MetricQueryResponse, error) {
	var raw struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`  // instant query
				Values [][]interface{}   `json:"values"` // range query
			} `json:"result"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, obserr.Wrap("PROMETHEUS_QUERY_FAILED", op, "failed to parse prometheus response", err)
	}
	if raw.Status == "error" {
		return nil, obserr.New("PROMETHEUS_QUERY_FAILED", op, raw.Error)
	}

	result := &model.MetricQueryResponse{ResultType: raw.Data.ResultType}
	for _, r := range raw.Data.Result {
		series := model.MetricSeries{Metric: r.Metric}
		if r.Value != nil && len(r.Value) == 2 {
			ts, _ := toFloat64(r.Value[0])
			val, _ := toFloat64(r.Value[1])
			series.Values = []model.MetricResult{{Timestamp: int64(ts), Value: val}}
		}
		for _, v := range r.Values {
			if len(v) == 2 {
				ts, _ := toFloat64(v[0])
				val, _ := toFloat64(v[1])
				series.Values = append(series.Values, model.MetricResult{Timestamp: int64(ts), Value: val})
			}
		}
		result.Results = append(result.Results, series)
	}
	return result, nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}
