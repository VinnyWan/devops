package service

import (
	"strings"
	"testing"
	"time"
)

func TestAlertServiceListRules_FilterKeyword(t *testing.T) {
	svc := NewAlertService()
	result := svc.ListRules("memory")
	if result.Total == 0 {
		t.Fatalf("expected matched rules")
	}
	for _, item := range result.Items {
		matched := strings.Contains(strings.ToLower(item.Name), "memory") || strings.Contains(strings.ToLower(item.Expr), "memory")
		if !matched {
			t.Fatalf("unexpected rule without keyword match: %s", item.Name)
		}
	}

	shortKeyword := svc.ListRules("me")
	if shortKeyword.Total != 3 {
		t.Fatalf("expected short keyword ignored, got %d", shortKeyword.Total)
	}

	clusterKeyword := svc.ListRules("prod-sh")
	if clusterKeyword.Total != 1 {
		t.Fatalf("expected cluster field keyword matched one rule, got %d", clusterKeyword.Total)
	}
}

func TestAlertServiceListHistory_FilterStatusAndRange(t *testing.T) {
	svc := NewAlertService()
	start := time.Now().Add(-time.Hour)
	end := time.Now()
	result := svc.ListHistory("firing", start, end)
	if result.Total != 1 {
		t.Fatalf("expected 1 firing alert in range, got %d", result.Total)
	}
	if result.Items[0].Status != "firing" {
		t.Fatalf("expected firing status, got %s", result.Items[0].Status)
	}
}

func TestAlertServiceSetRuleEnabled(t *testing.T) {
	svc := NewAlertService()
	rule, err := svc.SetRuleEnabled(RuleEnableRequest{ID: 3, Enabled: true})
	if err != nil {
		t.Fatalf("expected rule to be found: %v", err)
	}
	if !rule.Enabled {
		t.Fatalf("expected rule enabled")
	}
}

func TestAlertServiceUpsertSilenceAndFilter(t *testing.T) {
	svc := NewAlertService()
	start := time.Now().Add(10 * time.Minute)
	end := time.Now().Add(40 * time.Minute)
	created, err := svc.UpsertSilence(SilenceUpsertRequest{
		RuleID:    1,
		Reason:    "压测期间降噪",
		StartsAt:  start,
		EndsAt:    end,
		CreatedBy: "qa",
	})
	if err != nil {
		t.Fatalf("upsert silence failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected silence id")
	}
	filtered := svc.ListSilences(1)
	if filtered.Total == 0 {
		t.Fatalf("expected silence for rule 1")
	}
}

func TestAlertServiceUpsertChannelAndFilterByType(t *testing.T) {
	svc := NewAlertService()
	_, err := svc.UpsertChannel(ChannelUpsertRequest{
		Name:    "钉钉机器人",
		Type:    "dingtalk",
		Target:  "https://oapi.dingtalk.com/robot/send?access_token=fake",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("upsert channel failed: %v", err)
	}
	result := svc.ListChannels("dingtalk")
	if result.Total != 1 {
		t.Fatalf("expected one dingtalk channel, got %d", result.Total)
	}
}

func TestAlertServiceSaveConfig_InvalidEndpoint(t *testing.T) {
	svc := NewAlertService()
	_, err := svc.SaveConfig(SaveAlertmanagerConfigRequest{
		Endpoint: "http://invalid-alertmanager",
	})
	if err == nil {
		t.Fatalf("expected invalid endpoint error")
	}
}
