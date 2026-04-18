package model

import (
	"time"

	"gorm.io/gorm"
)

type CloudAccount struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TenantID      uint           `gorm:"not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:1" json:"tenantId"`
	Name          string         `gorm:"size:100;not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:2" json:"name"`
	Provider      string         `gorm:"size:20;not null;uniqueIndex:uk_cloud_tenant_provider_name,priority:3" json:"provider"`
	SecretID      string         `gorm:"size:500;not null" json:"-"`
	SecretKey     string         `gorm:"size:500;not null" json:"-"`
	Status        string         `gorm:"size:20;default:active" json:"status"`
	LastSyncAt    *time.Time     `json:"lastSyncAt"`
	LastSyncError string         `gorm:"type:text" json:"lastSyncError"`
	SyncInterval  int            `gorm:"default:60" json:"syncInterval"`
	Description   string         `gorm:"size:500" json:"description"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
