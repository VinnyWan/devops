package model

import (
	"time"

	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"` // 租户编码，用于数据隔离
	Description string `gorm:"size:255" json:"description"`
	Logo        string `gorm:"size:255" json:"logo"`
	Status      string `gorm:"size:20;default:'active';index" json:"status"` // active, inactive, suspended

	// 配额限制
	MaxUsers       int `gorm:"default:100" json:"maxUsers"`      // 最大用户数
	MaxDepartments int `gorm:"default:20" json:"maxDepartments"` // 最大部门数
	MaxRoles       int `gorm:"default:50" json:"maxRoles"`       // 最大角色数

	// 模块配置 (JSON格式)
	Modules string `gorm:"type:text" json:"modules"` // 启用的模块列表

	// SSO配置 (JSON格式)
	SSOConfig string `gorm:"type:text" json:"ssoConfig"` // 第三方SSO配置(飞书/钉钉/企微等)

	// 联系人
	ContactName  string `gorm:"size:100" json:"contactName"`
	ContactEmail string `gorm:"size:255" json:"contactEmail"`
	ContactPhone string `gorm:"size:50" json:"contactPhone"`

	// 时间戳
	ExpiresAt *time.Time     `json:"expiresAt"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}
