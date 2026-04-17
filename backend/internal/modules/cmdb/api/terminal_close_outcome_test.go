package api

import (
	"errors"
	"testing"

	cmdbterminal "devops-platform/internal/modules/cmdb/terminal"
)

func TestResolveTerminalCloseOutcome(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantStatus   string
		wantReason   string
		wantMessage  string
	}{
		{
			name:        "idle timeout",
			err:         cmdbterminal.ErrTerminalIdleTimeout,
			wantStatus:  "idle_timeout",
			wantReason:  "空闲超时自动断开",
			wantMessage: "Terminal idle timeout.",
		},
		{
			name:        "max duration",
			err:         cmdbterminal.ErrTerminalMaxDuration,
			wantStatus:  "max_duration",
			wantReason:  "会话时长超限自动断开",
			wantMessage: "Terminal max session duration exceeded.",
		},
		{
			name:        "normal close",
			err:         nil,
			wantStatus:  "closed",
			wantReason:  "用户主动关闭或连接正常结束",
			wantMessage: "Connection closed.",
		},
		{
			name:        "unexpected error",
			err:         errors.New("boom"),
			wantStatus:  "interrupted",
			wantReason:  "连接异常中断",
			wantMessage: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, reason, message := resolveTerminalCloseOutcome(tt.err)
			if status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", status, tt.wantStatus)
			}
			if reason != tt.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tt.wantReason)
			}
			if message != tt.wantMessage {
				t.Fatalf("message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}
