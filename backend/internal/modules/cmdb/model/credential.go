package model

import (
	"time"

	"gorm.io/gorm"
)

// Credential SSH 凭据
type Credential struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    *uint          `gorm:"index;uniqueIndex:uk_cmdb_credentials_tenant_name" json:"tenantId"`
	Name        string         `gorm:"size:100;not null;uniqueIndex:uk_cmdb_credentials_tenant_name" json:"name"`
	Type        string         `gorm:"size:20;not null" json:"type"` // password / key
	Username    string         `gorm:"size:100;not null" json:"username"`
	Password    string         `gorm:"type:text" json:"password"`     // AES-256 加密
	PrivateKey  string         `gorm:"type:text" json:"-"`            // AES-256 加密，API 不返回
	Passphrase  string         `gorm:"size:500" json:"-"`             // AES-256 加密，API 不返回
	Description string         `gorm:"size:500" json:"description"`
	CreatedAt   time.Time      `gorm:"index" json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
