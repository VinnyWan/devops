package model

import (
	"time"

	"gorm.io/gorm"
)

// HarborConfig holds Harbor registry connection info
type HarborConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	URL       string         `gorm:"size:512;not null" json:"url"`
	Username  string         `gorm:"size:128;not null" json:"username"`
	Password  string         `gorm:"size:256;not null" json:"-"`
	Status    string         `gorm:"size:20;default:'unknown'" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (HarborConfig) TableName() string { return "harbor_configs" }

// Project represents a Harbor project
type Project struct {
	ID           int    `json:"projectId"`
	Name         string `json:"name"`
	Public       bool   `json:"public"`
	RepoCount    int    `json:"repoCount"`
	RegistryID   *int   `json:"registryId,omitempty"`
	CreationTime string `json:"creationTime"`
}

// Repository represents an image repository in Harbor
type Repository struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProjectID    int    `json:"projectId"`
	ArtifactCount int   `json:"artifactCount"`
	PullCount    int    `json:"pullCount"`
	CreationTime string `json:"creationTime"`
	UpdateTime   string `json:"updateTime"`
}

// Artifact represents a specific image artifact
type Artifact struct {
	ID       int           `json:"id"`
	Digest   string        `json:"digest"`
	Size     int64         `json:"size"`
	PushTime string        `json:"pushTime"`
	PullTime string        `json:"pullTime"`
	Tags     []ArtifactTag `json:"tags"`
	Type     string        `json:"type"`
}

// ArtifactTag represents an image tag
type ArtifactTag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	PushTime  string `json:"pushTime"`
	PullTime  string `json:"pullTime"`
	Immutable bool   `json:"immutable"`
}

// ProjectSearchResult for search API
type ProjectSearchResult struct {
	Projects []Project `json:"projects"`
	Total    int64     `json:"total"`
}
