package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Notifier is the interface for sending notifications through a specific channel.
type Notifier interface {
	Send(recipients []string, subject, body string) error
}

// FeishuNotifier sends notifications via Feishu webhook.
type FeishuNotifier struct {
	WebhookURL string
}

func NewFeishuNotifier(webhookURL string) *FeishuNotifier {
	return &FeishuNotifier{WebhookURL: webhookURL}
}

func (n *FeishuNotifier) Send(recipients []string, subject, body string) error {
	msg := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title":    map[string]string{"content": subject},
			},
			"elements": []map[string]interface{}{
				{"tag": "markdown", "content": body},
			},
		},
	}
	return n.post(msg)
}

func (n *FeishuNotifier) post(msg interface{}) error {
	data, _ := json.Marshal(msg)
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("feishu webhook failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}
	return nil
}

// DingTalkNotifier sends notifications via DingTalk webhook.
type DingTalkNotifier struct {
	WebhookURL string
}

func NewDingTalkNotifier(webhookURL string) *DingTalkNotifier {
	return &DingTalkNotifier{WebhookURL: webhookURL}
}

func (n *DingTalkNotifier) Send(recipients []string, subject, body string) error {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": subject,
			"text":  fmt.Sprintf("## %s\n\n%s", subject, body),
		},
	}
	data, _ := json.Marshal(msg)
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("dingtalk webhook failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk returned status %d", resp.StatusCode)
	}
	return nil
}

// EmailNotifier sends notifications via SMTP.
type EmailNotifier struct {
	SMTPServer string
	Port       int
	Username   string
	Password   string
	From       string
}

func NewEmailNotifier(server string, port int, user, pass, from string) *EmailNotifier {
	return &EmailNotifier{
		SMTPServer: server,
		Port:       port,
		Username:   user,
		Password:   pass,
		From:       from,
	}
}

func (n *EmailNotifier) Send(recipients []string, subject, body string) error {
	// Placeholder: email sending requires net/smtp or a library like gomail
	// Full SMTP implementation deferred to when real email config is available
	_ = recipients
	_ = subject
	_ = body
	return nil
}
