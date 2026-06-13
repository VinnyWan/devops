package model

import (
	"time"

	"gorm.io/gorm"
)

// ToolStatus represents the installation state on a host.
type ToolStatus string

const (
	ToolNotInstalled ToolStatus = "not_installed"
	ToolInstalling   ToolStatus = "installing"
	ToolInstalled    ToolStatus = "installed"
	ToolFailed       ToolStatus = "failed"
)

// Tool defines an installable service/tool in the marketplace.
type Tool struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:64;not null;uniqueIndex" json:"name"`
	DisplayName    string         `gorm:"size:128;not null" json:"displayName"`
	Description    string         `gorm:"size:512" json:"description"`
	Category       string         `gorm:"size:32" json:"category"`
	Logo           string         `gorm:"size:256" json:"logo"`
	InstallScript  string         `gorm:"type:text;not null" json:"-"`
	CheckScript    string         `gorm:"type:text" json:"-"`
	VersionCmd     string         `gorm:"size:128" json:"versionCmd"`
	DefaultVersion string         `gorm:"size:32" json:"defaultVersion"`
	Dependencies   string         `gorm:"size:512" json:"dependencies"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Tool) TableName() string { return "tool_marketplace" }

// ToolInstallation records a tool installation on a specific host.
type ToolInstallation struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TenantID  uint       `gorm:"index;not null" json:"tenantId"`
	ToolID    uint       `gorm:"index;not null" json:"toolId"`
	HostID    uint       `gorm:"index;not null" json:"hostId"`
	HostIP    string     `gorm:"size:45;not null" json:"hostIp"`
	Version   string     `gorm:"size:32" json:"version"`
	Status    ToolStatus `gorm:"size:20;default:not_installed" json:"status"`
	Log       string     `gorm:"type:text" json:"log"`
	InstalledAt *time.Time `json:"installedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (ToolInstallation) TableName() string { return "tool_installations" }
