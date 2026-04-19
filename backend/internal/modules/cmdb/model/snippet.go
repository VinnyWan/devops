package model

import (
	"time"

	"gorm.io/gorm"
)

// CommandSnippet 命令片段
type CommandSnippet struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	TenantID   uint           `gorm:"not null;index:idx_cmdb_snippets_tenant" json:"tenantId"`
	UserID     uint           `gorm:"not null;index:idx_cmdb_snippets_user" json:"userId"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Tags       string         `gorm:"size:500" json:"tags"`
	Visibility string         `gorm:"size:20;not null;default:'personal'" json:"visibility"` // personal, team, public
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
