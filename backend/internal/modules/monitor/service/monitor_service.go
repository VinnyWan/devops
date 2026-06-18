package service

import (
	"fmt"

	"devops-platform/internal/modules/monitor/model"
	"devops-platform/internal/modules/monitor/repository"
	"devops-platform/internal/pkg/obserr"

	"gorm.io/gorm"
)

const op = "monitor/service"

// MonitorService provides business logic for Prometheus config and metric queries
type MonitorService struct {
	repo *repository.MonitorRepo
	db   *gorm.DB
}

// NewMonitorService creates a new MonitorService
func NewMonitorService(db *gorm.DB) *MonitorService {
	return &MonitorService{repo: repository.NewMonitorRepo(db), db: db}
}

// --- Config management ---

// ListConfigs returns paginated Prometheus configs
func (s *MonitorService) ListConfigs(page, pageSize int) ([]model.PrometheusConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListConfigs(page, pageSize)
}

// GetConfig returns a single Prometheus config by ID
func (s *MonitorService) GetConfig(id uint) (*model.PrometheusConfig, error) {
	return s.repo.GetConfig(id)
}

// SaveConfig validates and saves a Prometheus config, testing the connection first
func (s *MonitorService) SaveConfig(cfg *model.PrometheusConfig) error {
	if cfg.Endpoint == "" {
		return obserr.New("INVALID_PARAM", op, "endpoint is required")
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 15
	}
	if err := s.repo.TestConnection(cfg.Endpoint, cfg.Username, cfg.Password); err != nil {
		cfg.Status = "error"
	} else {
		cfg.Status = "connected"
	}
	return s.repo.SaveConfig(cfg)
}

// DeleteConfig soft-deletes a Prometheus config by ID
func (s *MonitorService) DeleteConfig(id uint) error {
	return s.repo.DeleteConfig(id)
}

// TestConnection tests connectivity to a Prometheus endpoint
func (s *MonitorService) TestConnection(endpoint, username, password string) error {
	if endpoint == "" {
		return obserr.New("INVALID_PARAM", op, "endpoint is required")
	}
	return s.repo.TestConnection(endpoint, username, password)
}

// --- Metric queries ---

// QueryHostMetrics builds a PromQL query for a given host metric and executes it
func (s *MonitorService) QueryHostMetrics(configID uint, hostIP, metric string, startTime, endTime string) (*model.MetricQueryResponse, error) {
	var promQL string
	switch metric {
	case "cpu":
		promQL = fmt.Sprintf(`100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle",instance=~"%s:.*"}[5m])) * 100)`, hostIP)
	case "memory":
		promQL = fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{instance=~"%s:.*"} / node_memory_MemTotal_bytes{instance=~"%s:.*"})) * 100`, hostIP, hostIP)
	case "disk":
		promQL = fmt.Sprintf(`100 - ((node_filesystem_avail_bytes{instance=~"%s:.*",mountpoint="/"} / node_filesystem_size_bytes{instance=~"%s:.*",mountpoint="/"}) * 100)`, hostIP, hostIP)
	case "disk_usage":
		promQL = fmt.Sprintf(`node_filesystem_size_bytes{instance=~"%s:.*",mountpoint="/"} - node_filesystem_avail_bytes{instance=~"%s:.*",mountpoint="/"}`, hostIP, hostIP)
	default:
		return nil, obserr.New("INVALID_PARAM", op, fmt.Sprintf("unsupported metric: %s", metric))
	}

	if startTime != "" && endTime != "" {
		step := "60s"
		return s.repo.QueryRange(configID, promQL, startTime, endTime, step)
	}
	return s.repo.QueryInstant(configID, promQL)
}

// QueryPortStatus queries port reachability via blackbox_exporter probe_success metric
func (s *MonitorService) QueryPortStatus(configID uint, hostIP string, ports []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, port := range ports {
		promQL := fmt.Sprintf(`probe_success{instance="%s:%s"}`, hostIP, port)
		resp, err := s.repo.QueryInstant(configID, promQL)
		if err != nil {
			result[port] = "unknown"
			continue
		}
		if len(resp.Results) > 0 && len(resp.Results[0].Values) > 0 {
			if resp.Results[0].Values[0].Value == 1 {
				result[port] = "up"
			} else {
				result[port] = "down"
			}
		} else {
			result[port] = "unknown"
		}
	}
	return result, nil
}

// QueryAgentStatus queries agent liveness via Prometheus up metric
func (s *MonitorService) QueryAgentStatus(configID uint, hostIPs []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, ip := range hostIPs {
		promQL := fmt.Sprintf(`up{instance=~"%s:.*"}`, ip)
		resp, err := s.repo.QueryInstant(configID, promQL)
		if err != nil {
			result[ip] = "unknown"
			continue
		}
		if len(resp.Results) > 0 && len(resp.Results[0].Values) > 0 {
			if resp.Results[0].Values[0].Value == 1 {
				result[ip] = "online"
			} else {
				result[ip] = "offline"
			}
		} else {
			result[ip] = "not_deployed"
		}
	}
	return result, nil
}

// EnsureDefaults seeds a default Prometheus config if the table is empty
func (s *MonitorService) EnsureDefaults() {
	var count int64
	s.db.Model(&model.PrometheusConfig{}).Count(&count)
	if count == 0 {
		s.db.Create(&model.PrometheusConfig{
			Name:           "default",
			Endpoint:       "http://prometheus:9090",
			TimeoutSeconds: 15,
			Status:         "unknown",
		})
	}
}
