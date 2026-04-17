package model

import (
	"time"

	"gorm.io/gorm"
)

// HostGroup 主机分组（三级固定层级：业务→环境→地域/机房）
type HostGroup struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  *uint          `gorm:"index;uniqueIndex:uk_cmdb_host_groups_tenant_parent_name" json:"tenantId"`
	Name      string         `gorm:"size:100;not null;uniqueIndex:uk_cmdb_host_groups_tenant_parent_name" json:"name"`
	Level     int            `gorm:"not null" json:"level"` // 1=业务, 2=环境, 3=地域/机房
	ParentID  uint           `gorm:"default:0;uniqueIndex:uk_cmdb_host_groups_tenant_parent_name" json:"parentId"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
