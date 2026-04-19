package model

import (
	"time"

	"gorm.io/gorm"
)

// SessionTag 终端会话标签
type SessionTag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"not null;index:idx_cmdb_session_tags_tenant" json:"tenantId"`
	SessionID uint           `gorm:"not null;index:idx_cmdb_session_tags_session" json:"sessionId"`
	Tag       string         `gorm:"size:50;not null;index:idx_cmdb_session_tags_tag" json:"tag"`
	UserID    uint           `gorm:"not null" json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
