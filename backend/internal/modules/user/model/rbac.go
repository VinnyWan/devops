package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type DataScope string

const (
	DataScopeSelfDepartment DataScope = "self_department"
	DataScopeDepartmentTree DataScope = "department_tree"
	DataScopeTenant         DataScope = "tenant"
)

func NormalizeDataScope(scope string) DataScope {
	switch DataScope(strings.TrimSpace(scope)) {
	case DataScopeTenant:
		return DataScopeTenant
	case DataScopeDepartmentTree:
		return DataScopeDepartmentTree
	default:
		return DataScopeSelfDepartment
	}
}

func IsValidDataScope(scope string) bool {
	switch DataScope(strings.TrimSpace(scope)) {
	case DataScopeSelfDepartment, DataScopeDepartmentTree, DataScopeTenant:
		return true
	default:
		return false
	}
}

func DefaultRoleDataScope(name string) (DataScope, bool) {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "SYSTEM_ADMIN", "TENANT_ADMIN":
		return DataScopeTenant, true
	case "DEPT_ADMIN":
		return DataScopeDepartmentTree, true
	case "READ_ONLY":
		return DataScopeSelfDepartment, true
	default:
		return DataScopeSelfDepartment, false
	}
}

func MaxDataScope(current DataScope, candidate DataScope) DataScope {
	if dataScopeRank(candidate) > dataScopeRank(current) {
		return candidate
	}
	return current
}

func dataScopeRank(scope DataScope) int {
	switch NormalizeDataScope(string(scope)) {
	case DataScopeTenant:
		return 3
	case DataScopeDepartmentTree:
		return 2
	default:
		return 1
	}
}

// Role 角色模型
type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	TenantID    *uint          `gorm:"index;uniqueIndex:uk_roles_tenant_name" json:"tenantId"` // 租户ID，空表示全局角色
	Name        string         `gorm:"size:50;not null;uniqueIndex:uk_roles_tenant_name" json:"name"`
	DisplayName string         `gorm:"column:display_name;size:100" json:"displayName"`
	Description string         `gorm:"size:200" json:"description"`
	Type        string         `gorm:"size:20;default:'custom'" json:"type"` // system (内置), custom (自定义)
	DataScope   DataScope      `gorm:"column:data_scope;size:32;not null;default:'self_department'" json:"dataScope"`
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
