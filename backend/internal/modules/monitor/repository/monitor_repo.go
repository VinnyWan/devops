package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/monitor/model"
)

type MonitorRepo struct {
	mu     sync.RWMutex
	config model.PrometheusConfig
}

func NewMonitorRepo() *MonitorRepo {
	return &MonitorRepo{
		config: model.PrometheusConfig{
			Endpoint:       "http://prometheus.monitoring.svc:9090",
			QueryPath:      "/api/v1/query_range",
			TimeoutSeconds: 10,
			UpdatedAt:      time.Now(),
		},
	}
}

func (r *MonitorRepo) Query(req model.QueryRequest) model.QueryResult {
	baseStart := req.Start.Truncate(time.Minute)
	return model.QueryResult{
		Metric: req.Metric,
		Start:  req.Start,
		End:    req.End,
		Step:   req.Step,
		Series: []model.QuerySeries{
			{
				Labels: map[string]string{
					"cluster":   "default",
					"namespace": "kube-system",
					"instance":  "node-1",
				},
				Points: []model.QueryPoint{
					{Timestamp: baseStart, Value: 0.42},
					{Timestamp: baseStart.Add(time.Minute), Value: 0.48},
					{Timestamp: baseStart.Add(2 * time.Minute), Value: 0.51},
				},
			},
		},
	}
}

func (r *MonitorRepo) GetConfig() model.PrometheusConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

func (r *MonitorRepo) SaveConfig(cfg model.PrometheusConfig) model.PrometheusConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg.UpdatedAt = time.Now()
	r.config = cfg
	return r.config
}

func (r *MonitorRepo) ValidateConfigConnection(cfg model.PrometheusConfig) error {
	if strings.Contains(strings.ToLower(cfg.Endpoint), "invalid") {
		return errors.New("prometheus endpoint 不可达")
	}
	if strings.Contains(strings.ToLower(cfg.Endpoint), "timeout") {
		return errors.New("prometheus 请求超时")
	}
	return nil
}
