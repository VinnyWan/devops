package model

import (
	"time"

	"gorm.io/gorm"
)

type CloudResource struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TenantID       uint           `gorm:"not null;index" json:"tenantId"`
	CloudAccountID uint           `gorm:"not null;uniqueIndex:uk_cloud_res,priority:1;index" json:"cloudAccountId"`
	ResourceType   string         `gorm:"size:30;not null;uniqueIndex:uk_cloud_res,priority:2;index:idx_cloud_res_type" json:"resourceType"`
	ResourceID     string         `gorm:"size:100;not null;uniqueIndex:uk_cloud_res,priority:3" json:"resourceId"`
	Region         string         `gorm:"size:50;index:idx_cloud_res_region" json:"region"`
	Zone           string         `gorm:"size:50" json:"zone"`
	Name           string         `gorm:"size:200" json:"name"`
	State          string         `gorm:"size:30" json:"state"`
	Spec           string         `gorm:"type:text" json:"spec"`
	SyncedAt       time.Time      `json:"syncedAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
