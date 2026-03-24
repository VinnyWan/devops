package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"devops-platform/internal/modules/alert/model"
)

type AlertRepo struct {
	mu       sync.RWMutex
	nextID   uint
	rules    []model.Rule
	silences []model.Silence
	channels []model.NotificationChannel
	config   model.AlertmanagerConfig
}

func NewAlertRepo() *AlertRepo {
	now := time.Now()
	return &AlertRepo{
		nextID: 100,
		rules: []model.Rule{
			{
				ID:          1,
				Name:        "PodCrashLooping",
				Expr:        `sum(rate(kube_pod_container_status_restarts_total[5m])) > 0`,
				Severity:    "warning",
				Enabled:     true,
				Cluster:     "default",
				UpdatedAt:   now.Add(-30 * time.Minute),
				Description: "容器重启频繁",
			},
			{
				ID:          2,
				Name:        "NodeMemoryHigh",
				Expr:        `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) > 0.9`,
				Severity:    "critical",
				Enabled:     true,
				Cluster:     "prod-sh",
				UpdatedAt:   now.Add(-2 * time.Hour),
				Description: "节点内存使用率过高",
			},
			{
				ID:          3,
				Name:        "ApiServerLatencyP99",
				Expr:        `histogram_quantile(0.99, sum(rate(apiserver_request_duration_seconds_bucket[5m])) by (le)) > 1`,
				Severity:    "warning",
				Enabled:     false,
				Cluster:     "prod-bj",
				UpdatedAt:   now.Add(-24 * time.Hour),
				Description: "apiserver 延迟过高",
			},
		},
		silences: []model.Silence{
			{
				ID:        11,
				RuleID:    2,
				Reason:    "生产变更窗口",
				StartsAt:  now.Add(-10 * time.Minute),
				EndsAt:    now.Add(50 * time.Minute),
				CreatedBy: "sre",
				UpdatedAt: now.Add(-10 * time.Minute),
			},
		},
		channels: []model.NotificationChannel{
			{
				ID:        21,
				Name:      "值班邮箱",
				Type:      "email",
				Target:    "sre@example.com",
				Enabled:   true,
				UpdatedAt: now.Add(-2 * time.Hour),
			},
			{
				ID:        22,
				Name:      "报警机器人",
				Type:      "slack",
				Target:    "#prod-alert",
				Enabled:   true,
				UpdatedAt: now.Add(-30 * time.Minute),
			},
		},
		config: model.AlertmanagerConfig{
			Endpoint:       "http://alertmanager.monitoring.svc:9093",
			APIPath:        "/api/v2/alerts",
			TimeoutSeconds: 10,
			UpdatedAt:      now,
		},
	}
}

func (r *AlertRepo) ListRules() []model.Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Rule(nil), r.rules...)
}

func (r *AlertRepo) SetRuleEnabled(id uint, enabled bool) (model.Rule, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.rules {
		if r.rules[i].ID != id {
			continue
		}
		r.rules[i].Enabled = enabled
		r.rules[i].UpdatedAt = time.Now()
		return r.rules[i], true
	}
	return model.Rule{}, false
}

func (r *AlertRepo) ListSilences() []model.Silence {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Silence(nil), r.silences...)
}

func (r *AlertRepo) UpsertSilence(silence model.Silence) model.Silence {
	r.mu.Lock()
	defer r.mu.Unlock()
	silence.UpdatedAt = time.Now()
	if silence.ID > 0 {
		for i := range r.silences {
			if r.silences[i].ID == silence.ID {
				r.silences[i] = silence
				return silence
			}
		}
	}
	r.nextID++
	silence.ID = r.nextID
	r.silences = append(r.silences, silence)
	return silence
}

func (r *AlertRepo) ListChannels() []model.NotificationChannel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.NotificationChannel(nil), r.channels...)
}

func (r *AlertRepo) UpsertChannel(channel model.NotificationChannel) model.NotificationChannel {
	r.mu.Lock()
	defer r.mu.Unlock()
	channel.UpdatedAt = time.Now()
	if channel.ID > 0 {
		for i := range r.channels {
			if r.channels[i].ID == channel.ID {
				r.channels[i] = channel
				return channel
			}
		}
	}
	r.nextID++
	channel.ID = r.nextID
	r.channels = append(r.channels, channel)
	return channel
}

func (r *AlertRepo) ListHistory() []model.History {
	now := time.Now()
	return []model.History{
		{
			ID:          1001,
			RuleID:      2,
			RuleName:    "NodeMemoryHigh",
			Status:      "firing",
			Severity:    "critical",
			Summary:     "node-1 内存使用率达到 92%",
			StartsAt:    now.Add(-20 * time.Minute),
			EndsAt:      time.Time{},
			Cluster:     "prod-sh",
			Namespace:   "kube-system",
			Instance:    "node-1",
			Fingerprint: "hist-1001",
		},
		{
			ID:          1002,
			RuleID:      1,
			RuleName:    "PodCrashLooping",
			Status:      "resolved",
			Severity:    "warning",
			Summary:     "payments-api 重启已恢复",
			StartsAt:    now.Add(-3 * time.Hour),
			EndsAt:      now.Add(-2 * time.Hour),
			Cluster:     "default",
			Namespace:   "payments",
			Instance:    "payments-api-6f4f7d47f9-abcde",
			Fingerprint: "hist-1002",
		},
		{
			ID:          1003,
			RuleID:      3,
			RuleName:    "ApiServerLatencyP99",
			Status:      "resolved",
			Severity:    "warning",
			Summary:     "apiserver P99 延迟恢复到 800ms",
			StartsAt:    now.Add(-8 * time.Hour),
			EndsAt:      now.Add(-7 * time.Hour),
			Cluster:     "prod-bj",
			Namespace:   "kube-system",
			Instance:    "apiserver-1",
			Fingerprint: "hist-1003",
		},
	}
}

func (r *AlertRepo) GetConfig() model.AlertmanagerConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

func (r *AlertRepo) SaveConfig(cfg model.AlertmanagerConfig) model.AlertmanagerConfig {
	r.mu.Lock()
	defer r.mu.Unlock()
	cfg.UpdatedAt = time.Now()
	r.config = cfg
	return r.config
}

func (r *AlertRepo) ValidateConfigConnection(cfg model.AlertmanagerConfig) error {
	if strings.Contains(strings.ToLower(cfg.Endpoint), "invalid") {
		return errors.New("alertmanager endpoint 不可达")
	}
	if strings.Contains(strings.ToLower(cfg.Endpoint), "timeout") {
		return errors.New("alertmanager 请求超时")
	}
	return nil
}
