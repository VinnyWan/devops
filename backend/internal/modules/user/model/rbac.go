package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    *uint          `gorm:"index:idx_tenant_role_name" json:"tenantId"` // 租户ID，空表示全局角色
	Name        string         `gorm:"size:50;not null;index:idx_tenant_role_name" json:"name"`
	DisplayName string         `gorm:"column:display_name;size:100" json:"displayName"`
	Description string         `gorm:"size:200" json:"description"`
	Type        string         `gorm:"size:20;default:'custom'" json:"type"` // system (内置), custom (自定义)
	Tenant      *Tenant        `json:"tenant,omitempty"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions"`
	Departments []Department   `gorm:"many2many:department_roles;" json:"departments,omitempty"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Permission 权限模型
type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Resource    string         `gorm:"size:50;not null" json:"resource"` // e.g., "cluster", "user"
	Action      string         `gorm:"size:50;not null" json:"action"`   // e.g., "create", "read", "update", "delete"
	Description string         `gorm:"size:200" json:"description"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}
