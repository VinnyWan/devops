package model

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTerminalSessionJSONIncludesCloseReason(t *testing.T) {
	session := TerminalSession{
		ID:           1,
		TenantID:     2,
		UserID:       3,
		Username:     "alice",
		HostID:       4,
		HostIP:       "192.0.2.10",
		HostName:     "host-01",
		CredentialID: 5,
		ClientIP:     "198.51.100.20",
		StartedAt:    time.Date(2026, 4, 17, 8, 0, 0, 0, time.UTC),
		Status:       "closed",
		CloseReason:  "空闲超时自动断开",
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	got := string(data)
	if !strings.Contains(got, `"closeReason":"空闲超时自动断开"`) {
		t.Fatalf("marshaled JSON %q does not include closeReason", got)
	}
}
