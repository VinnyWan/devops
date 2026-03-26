package model

import (
	"time"

	"gorm.io/gorm"
)

// AuthType 认证类型
type AuthType string

const (
	AuthTypeLocal AuthType = "local" // 内建用户库
	AuthTypeLDAP  AuthType = "ldap"  // LDAP认证
	AuthTypeOIDC  AuthType = "oidc"  // OIDC认证
)

// User 用户模型
type User struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   *uint  `gorm:"index" json:"tenantId"` // 租户ID
	Username   string `gorm:"size:100;not null;index:idx_tenant_username" json:"username"`
	Password   string `gorm:"size:255" json:"-"`
	Email      string `gorm:"size:255;index:idx_tenant_email" json:"email"`
	Name       string `gorm:"size:100;index" json:"name"`
	ExternalID string `gorm:"size:255" json:"externalId"`

	AuthType AuthType `gorm:"size:20;default:'local';index" json:"authType"`

	Status   string `gorm:"size:20;default:'active';index" json:"status"`
	IsAdmin  bool   `gorm:"default:false" json:"isAdmin"`
	IsLocked bool   `gorm:"default:false" json:"isLocked"`

	DepartmentID *uint       `gorm:"index" json:"departmentId"`
	Department   *Department `json:"department,omitempty"`
	Tenant       *Tenant     `json:"tenant,omitempty"`

	Roles []Role `gorm:"many2many:user_roles;" json:"roles"`

	LastLoginAt *time.Time     `json:"lastLoginAt"`
	CreatedAt   time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TenantUser 租户用户关联 (用于查询)
type TenantUser struct {
	TenantID uint
	UserID   uint
}
