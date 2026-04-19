package model

import (
	"time"

	"gorm.io/gorm"
)

// TerminalSession 主机终端会话
// 用于审计远程终端连接、录屏文件及会话状态。
type TerminalSession struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TenantID      uint           `gorm:"not null;index:idx_cmdb_terminal_tenant_started,priority:1;index:idx_cmdb_terminal_tenant_host,priority:1;index:idx_cmdb_terminal_tenant_user,priority:1;index:idx_cmdb_terminal_tenant_status,priority:1" json:"tenantId"`
	UserID        uint           `gorm:"not null;index:idx_cmdb_terminal_tenant_user,priority:2" json:"userId"`
	Username      string         `gorm:"size:100;not null" json:"username"`
	HostID        uint           `gorm:"not null;index:idx_cmdb_terminal_tenant_host,priority:2" json:"hostId"`
	HostIP        string         `gorm:"size:45;not null" json:"hostIp"`
	HostName      string         `gorm:"size:255;not null" json:"hostName"`
	CredentialID  uint           `gorm:"not null" json:"credentialId"`
	ClientIP      string         `gorm:"size:45" json:"clientIp"`
	StartedAt     time.Time      `gorm:"not null;index:idx_cmdb_terminal_tenant_started,priority:2" json:"startedAt"`
	FinishedAt    *time.Time     `json:"finishedAt"`
	Duration      int            `gorm:"default:0" json:"duration"`
	RecordingPath string         `gorm:"size:500;not null" json:"-"`
	FileSize      int64          `gorm:"default:0" json:"fileSize"`
	Status        string         `gorm:"size:20;not null;default:'active';index:idx_cmdb_terminal_tenant_status,priority:2" json:"status"`
	CloseReason    string         `gorm:"size:100" json:"closeReason"`
	Tags           string         `gorm:"size:500" json:"tags"`
	CommandSummary string         `gorm:"type:text" json:"commandSummary,omitempty"`
	CreatedAt     time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
