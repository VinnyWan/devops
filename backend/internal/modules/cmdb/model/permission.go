package model

import (
	"time"

	"gorm.io/gorm"
)

type HostPermission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:1" json:"tenantId"`
	UserID      uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:2;index:idx_perm_user" json:"userId"`
	HostGroupID uint           `gorm:"not null;uniqueIndex:uk_perm_user_group,priority:3;index:idx_perm_group" json:"hostGroupId"`
	Permission  string         `gorm:"size:20;not null;uniqueIndex:uk_perm_user_group,priority:4" json:"permission"`
	CreatedBy   uint           `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
