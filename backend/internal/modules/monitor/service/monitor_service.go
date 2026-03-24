package service

import (
	"fmt"
	"strings"
	"time"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/modules/monitor/repository"
	"devops-platform/internal/pkg/obserr"
)

type MonitorService struct {
	repo *repository.MonitorRepo
}

type SavePrometheusConfigRequest struct {
	Endpoint              string `json:"endpoint"`
	QueryPath             string `json:"queryPath"`
	TimeoutSeconds        int    `json:"timeoutSeconds"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	BearerToken           string `json:"bearerToken"`
	TLSInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify"`
}

func NewMonitorService() *MonitorService {
	return &MonitorService{repo: repository.NewMonitorRepo()}
}

func (s *MonitorService) Query(metric string, start, end time.Time, step string) (model.QueryResult, error) {
	cfg := s.repo.GetConfig()
	if err := s.repo.ValidateConfigConnection(cfg); err != nil {
		return model.QueryResult{}, obserr.Wrap("PROMETHEUS_CONNECT_FAILED", "monitor.Query", "Prometheus 配置连接校验失败", err)
	}
	metric = strings.TrimSpace(metric)
	if metric == "" {
		metric = "cpu_usage"
	}
	if start.IsZero() {
		start = time.Now().Add(-15 * time.Minute)
	}
	if end.IsZero() || end.Before(start) {
		end = start.Add(15 * time.Minute)
	}
	step = strings.TrimSpace(step)
	if step == "" {
		step = "1m"
	}
	if !strings.HasSuffix(step, "s") && !strings.HasSuffix(step, "m") && !strings.HasSuffix(step, "h") {
		return model.QueryResult{}, obserr.New("PROMETHEUS_INVALID_STEP", "monitor.Query", fmt.Sprintf("无效的 step: %s", step))
	}
	result := s.repo.Query(model.QueryRequest{
		Metric: metric,
		Start:  start,
		End:    end,
		Step:   step,
	})
	return result, nil
}

func (s *MonitorService) GetConfig() model.PrometheusConfig {
	return s.repo.GetConfig()
}

func (s *MonitorService) SaveConfig(req SavePrometheusConfigRequest) (model.PrometheusConfig, error) {
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		return model.PrometheusConfig{}, obserr.New("PROMETHEUS_ENDPOINT_REQUIRED", "monitor.SaveConfig", "Prometheus endpoint 不能为空")
	}
	queryPath := strings.TrimSpace(req.QueryPath)
	if queryPath == "" {
		queryPath = "/api/v1/query_range"
	}
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	config := model.PrometheusConfig{
		Endpoint:              endpoint,
		QueryPath:             queryPath,
		TimeoutSeconds:        timeout,
		Username:              strings.TrimSpace(req.Username),
		Password:              req.Password,
		BearerToken:           strings.TrimSpace(req.BearerToken),
		TLSInsecureSkipVerify: req.TLSInsecureSkipVerify,
	}
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return model.PrometheusConfig{}, obserr.Wrap("PROMETHEUS_CONNECT_FAILED", "monitor.SaveConfig", "Prometheus 配置连接失败", err)
	}
	saved := s.repo.SaveConfig(config)
	return saved, nil
}

func (s *MonitorService) ValidateCurrentConfig() error {
	config := s.repo.GetConfig()
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return obserr.Wrap("PROMETHEUS_CONNECT_FAILED", "monitor.ValidateCurrentConfig", "Prometheus 配置连接失败", err)
	}
	return nil
}
