package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"index" json:"userId"`
	Username      string         `gorm:"size:50;index" json:"username"`
	Operation     string         `gorm:"size:100" json:"operation"`
	Method        string         `gorm:"size:10" json:"method"`
	Path          string         `gorm:"size:200;index" json:"path"`
	Params        string         `gorm:"type:text" json:"params"`
	Result        string         `gorm:"type:text" json:"result"`
	ErrorMessage  string         `gorm:"type:text" json:"errorMessage"`
	IP            string         `gorm:"size:50" json:"ip"`
	Status        int            `json:"status"`
	Latency       int64          `json:"latency"`
	RetentionDays int            `gorm:"default:3" json:"retentionDays"`
	RequestAt     time.Time      `gorm:"index" json:"requestAt"`
	CreatedAt     time.Time      `gorm:"index" json:"createdAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
