package model

import (
	"time"

	"gorm.io/gorm"
)

// FileOperationLog 记录文件操作审计日志
type FileOperationLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"index;not null" json:"tenantId"`
	UserID    uint           `gorm:"index;not null" json:"userId"`
	Username  string         `gorm:"size:100" json:"username"`
	HostID    uint           `gorm:"index;not null" json:"hostId"`
	HostIP    string         `gorm:"size:45" json:"hostIp"`
	HostName  string         `gorm:"size:200" json:"hostName"`
	OpType    string         `gorm:"size:20;not null;index" json:"opType"`
	FilePath  string         `gorm:"size:500;not null;index" json:"filePath"`
	FileSize  int64          `json:"fileSize"`
	Result    string         `gorm:"size:20;not null;index" json:"result"`
	ErrorMsg  string         `gorm:"type:text" json:"errorMsg,omitempty"`
	ClientIP  string         `gorm:"size:45" json:"clientIp"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FileOperationLog) TableName() string {
	return "cmdb_file_audit_log"
}
