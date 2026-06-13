package service

import (
	"fmt"
	"strings"

	"devops-platform/internal/modules/alert/model"
	notifModel "devops-platform/internal/modules/notification/model"
	notifService "devops-platform/internal/modules/notification/service"
)

// AlertNotificationBridge wires alert module events to the notification hub.
type AlertNotificationBridge struct {
	notificationService *notifService.NotificationService
}

func NewAlertNotificationBridge(ns *notifService.NotificationService) *AlertNotificationBridge {
	return &AlertNotificationBridge{notificationService: ns}
}

// AlertInfo carries the data needed to send an alert notification.
type AlertInfo struct {
	RuleName    string
	Severity    string
	Cluster     string
	Expr        string
	Status      string
	Description string
	Labels      map[string]string
}

// SendAlert sends an alert via the notification hub.
// It formats the alert body using the template engine and dispatches to all configured channels with fallback.
func (b *AlertNotificationBridge) SendAlert(tenantID uint, alert AlertInfo) error {
	subject := fmt.Sprintf("[%s] %s - %s", strings.ToUpper(alert.Severity), alert.Status, alert.RuleName)
	body := b.formatAlertBody(alert)

	return b.notificationService.SendWithFallback(tenantID, nil, subject, body)
}

func (b *AlertNotificationBridge) formatAlertBody(alert AlertInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**规则名称:** %s\n\n", alert.RuleName))
	sb.WriteString(fmt.Sprintf("**级别:** %s\n\n", alert.Severity))
	sb.WriteString(fmt.Sprintf("**状态:** %s\n\n", alert.Status))
	if alert.Cluster != "" {
		sb.WriteString(fmt.Sprintf("**集群:** %s\n\n", alert.Cluster))
	}
	if alert.Expr != "" {
		sb.WriteString(fmt.Sprintf("**表达式:** %s\n\n", alert.Expr))
	}
	if alert.Description != "" {
		sb.WriteString(fmt.Sprintf("**描述:** %s\n\n", alert.Description))
	}
	if len(alert.Labels) > 0 {
		sb.WriteString("**附加标签:**\n")
		for k, v := range alert.Labels {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
	}
	return sb.String()
}

// HandleAlertRuleTriggered is called when an alert rule fires.
func (b *AlertNotificationBridge) HandleAlertRuleTriggered(tenantID uint, rule model.Rule, status string) error {
	return b.SendAlert(tenantID, AlertInfo{
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Cluster:     rule.Cluster,
		Expr:        rule.Expr,
		Status:      status,
		Description: rule.Description,
	})
}

// NotifierAdapter adapts the notification hub's notifier interface to the alert module's channel type.
func NotifierAdapter(alertChannelType string) notifModel.ChannelType {
	switch strings.ToLower(alertChannelType) {
	case "feishu":
		return notifModel.ChannelFeishu
	case "dingtalk":
		return notifModel.ChannelDingTalk
	case "wecom":
		return notifModel.ChannelWeCom
	case "email":
		return notifModel.ChannelEmail
	default:
		return notifModel.ChannelFeishu
	}
}
