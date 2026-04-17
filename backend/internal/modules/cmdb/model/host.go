package model

import (
	"time"

	"gorm.io/gorm"
)

// Host 主机资产
type Host struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TenantID        *uint          `gorm:"index;uniqueIndex:uk_cmdb_hosts_tenant_ip_port" json:"tenantId"`
	Hostname        string         `gorm:"size:255;not null" json:"hostname"`
	Ip              string         `gorm:"size:45;not null;uniqueIndex:uk_cmdb_hosts_tenant_ip_port" json:"ip"`
	Port            int            `gorm:"default:22;uniqueIndex:uk_cmdb_hosts_tenant_ip_port" json:"port"`
	OsType          string         `gorm:"size:20" json:"osType"`
	OsName          string         `gorm:"size:255" json:"osName"`
	CpuCores        int            `gorm:"default:0" json:"cpuCores"`
	MemoryTotal     int            `gorm:"default:0" json:"memoryTotal"`
	DiskTotal       int            `gorm:"default:0" json:"diskTotal"`
	Status          string         `gorm:"size:20;default:'unknown';index" json:"status"`
	CredentialID    *uint          `gorm:"index" json:"credentialId"`
	GroupID         *uint          `gorm:"index" json:"groupId"`
	CloudAccountID  *uint          `json:"cloudAccountId"`
	CloudInstanceID string         `gorm:"size:100" json:"cloudInstanceId"`
	Labels          string         `gorm:"size:500" json:"labels"`
	Description     string         `gorm:"size:500" json:"description"`
	AgentVersion    string         `gorm:"size:50" json:"agentVersion"`
	LastActiveAt    *time.Time     `json:"lastActiveAt"`
	CreatedAt       time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
