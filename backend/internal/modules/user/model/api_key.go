package model

import (
	"time"

	"gorm.io/gorm"
)

type ApiKey struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	TenantID  uint           `gorm:"index;not null" json:"tenant_id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	KeyHash   string         `gorm:"size:256;not null;uniqueIndex" json:"-"`
	KeyPrefix string         `gorm:"size:16;not null" json:"key_prefix"`
	Scopes    string         `gorm:"size:1024;default:''" json:"scopes"`
	ExpiresAt *time.Time     `json:"expires_at"`
	LastUsed  *time.Time     `json:"last_used"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

type CreateApiKeyRequest struct {
	Name     string `json:"name" binding:"required,max=128"`
	ExpireDays int  `json:"expire_days" binding:"min=0,max=3650"`
	Scopes   []string `json:"scopes"`
}

type ApiKeyResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	Scopes    string     `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
	LastUsed  *time.Time `json:"last_used"`
	CreatedAt time.Time  `json:"created_at"`
	Key       string     `json:"key,omitempty"`
}

type ListApiKeyRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
}
