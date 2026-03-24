package service

import (
	"errors"
	"strings"
	"time"

	"devops-platform/internal/modules/alert/model"
	"devops-platform/internal/modules/alert/repository"
	"devops-platform/internal/pkg/obserr"
	queryutil "devops-platform/internal/pkg/query"
)

type AlertService struct {
	repo *repository.AlertRepo
}

type ListRulesResponse struct {
	Total int          `json:"total"`
	Items []model.Rule `json:"items"`
}

type ListHistoryResponse struct {
	Total int             `json:"total"`
	Items []model.History `json:"items"`
}

type ListSilenceResponse struct {
	Total int             `json:"total"`
	Items []model.Silence `json:"items"`
}

type ListChannelResponse struct {
	Total int                         `json:"total"`
	Items []model.NotificationChannel `json:"items"`
}

type RuleEnableRequest struct {
	ID      uint `json:"id"`
	Enabled bool `json:"enabled"`
}

type SilenceUpsertRequest struct {
	ID        uint      `json:"id"`
	RuleID    uint      `json:"ruleId"`
	Reason    string    `json:"reason"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	CreatedBy string    `json:"createdBy"`
}

type ChannelUpsertRequest struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

type SaveAlertmanagerConfigRequest struct {
	Endpoint              string `json:"endpoint"`
	APIPath               string `json:"apiPath"`
	TimeoutSeconds        int    `json:"timeoutSeconds"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	BearerToken           string `json:"bearerToken"`
	TLSInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify"`
}

func NewAlertService() *AlertService {
	return &AlertService{repo: repository.NewAlertRepo()}
}

func (s *AlertService) ListRules(keyword string) ListRulesResponse {
	rules := s.repo.ListRules()
	items := make([]model.Rule, 0, len(rules))
	for _, rule := range rules {
		if !queryutil.MatchKeywordAny(keyword, rule.Name, rule.Expr, rule.Cluster, rule.Severity, rule.Description) {
			continue
		}
		items = append(items, rule)
	}
	return ListRulesResponse{Total: len(items), Items: items}
}

func (s *AlertService) ListHistory(status string, start, end time.Time) ListHistoryResponse {
	status = strings.TrimSpace(strings.ToLower(status))
	history := s.repo.ListHistory()
	items := make([]model.History, 0, len(history))
	for _, item := range history {
		if status != "" && strings.ToLower(item.Status) != status {
			continue
		}
		if !start.IsZero() && item.StartsAt.Before(start) {
			continue
		}
		if !end.IsZero() && item.StartsAt.After(end) {
			continue
		}
		items = append(items, item)
	}
	return ListHistoryResponse{Total: len(items), Items: items}
}

func (s *AlertService) SetRuleEnabled(req RuleEnableRequest) (model.Rule, error) {
	rule, ok := s.repo.SetRuleEnabled(req.ID, req.Enabled)
	if !ok {
		return model.Rule{}, obserr.New("ALERT_RULE_NOT_FOUND", "alert.SetRuleEnabled", "告警规则不存在")
	}
	return rule, nil
}

func (s *AlertService) ListSilences(ruleID uint) ListSilenceResponse {
	silences := s.repo.ListSilences()
	if ruleID == 0 {
		return ListSilenceResponse{Total: len(silences), Items: silences}
	}
	items := make([]model.Silence, 0, len(silences))
	for _, silence := range silences {
		if silence.RuleID == ruleID {
			items = append(items, silence)
		}
	}
	return ListSilenceResponse{Total: len(items), Items: items}
}

func (s *AlertService) UpsertSilence(req SilenceUpsertRequest) (model.Silence, error) {
	if req.RuleID == 0 {
		return model.Silence{}, obserr.New("ALERT_SILENCE_RULE_REQUIRED", "alert.UpsertSilence", "ruleId 不能为空")
	}
	if req.StartsAt.IsZero() || req.EndsAt.IsZero() || !req.EndsAt.After(req.StartsAt) {
		return model.Silence{}, obserr.New("ALERT_SILENCE_RANGE_INVALID", "alert.UpsertSilence", "静默时间范围无效")
	}
	return s.repo.UpsertSilence(model.Silence{
		ID:        req.ID,
		RuleID:    req.RuleID,
		Reason:    req.Reason,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
		CreatedBy: req.CreatedBy,
	}), nil
}

func (s *AlertService) ListChannels(channelType string) ListChannelResponse {
	channelType = strings.TrimSpace(strings.ToLower(channelType))
	channels := s.repo.ListChannels()
	if channelType == "" {
		return ListChannelResponse{Total: len(channels), Items: channels}
	}
	items := make([]model.NotificationChannel, 0, len(channels))
	for _, channel := range channels {
		if strings.ToLower(channel.Type) == channelType {
			items = append(items, channel)
		}
	}
	return ListChannelResponse{Total: len(items), Items: items}
}

func (s *AlertService) UpsertChannel(req ChannelUpsertRequest) (model.NotificationChannel, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Target) == "" {
		return model.NotificationChannel{}, obserr.New("ALERT_CHANNEL_INVALID", "alert.UpsertChannel", "通知渠道参数不完整")
	}
	return s.repo.UpsertChannel(model.NotificationChannel{
		ID:      req.ID,
		Name:    req.Name,
		Type:    req.Type,
		Target:  req.Target,
		Enabled: req.Enabled,
	}), nil
}

func (s *AlertService) GetConfig() model.AlertmanagerConfig {
	return s.repo.GetConfig()
}

func (s *AlertService) SaveConfig(req SaveAlertmanagerConfigRequest) (model.AlertmanagerConfig, error) {
	endpoint := strings.TrimSpace(req.Endpoint)
	if endpoint == "" {
		return model.AlertmanagerConfig{}, obserr.New("ALERTMANAGER_ENDPOINT_REQUIRED", "alert.SaveConfig", "Alertmanager endpoint 不能为空")
	}
	apiPath := strings.TrimSpace(req.APIPath)
	if apiPath == "" {
		apiPath = "/api/v2/alerts"
	}
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	config := model.AlertmanagerConfig{
		Endpoint:              endpoint,
		APIPath:               apiPath,
		TimeoutSeconds:        timeout,
		Username:              strings.TrimSpace(req.Username),
		Password:              req.Password,
		BearerToken:           strings.TrimSpace(req.BearerToken),
		TLSInsecureSkipVerify: req.TLSInsecureSkipVerify,
	}
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return model.AlertmanagerConfig{}, obserr.Wrap("ALERTMANAGER_CONNECT_FAILED", "alert.SaveConfig", "Alertmanager 配置连接失败", err)
	}
	saved := s.repo.SaveConfig(config)
	return saved, nil
}

func (s *AlertService) ValidateCurrentConfig() error {
	config := s.repo.GetConfig()
	if err := s.repo.ValidateConfigConnection(config); err != nil {
		return obserr.Wrap("ALERTMANAGER_CONNECT_FAILED", "alert.ValidateCurrentConfig", "Alertmanager 配置连接失败", err)
	}
	return nil
}

func (s *AlertService) IsNotFound(err error) bool {
	var observable *obserr.ObservableError
	return errors.As(err, &observable) && observable.Code == "ALERT_RULE_NOT_FOUND"
}
