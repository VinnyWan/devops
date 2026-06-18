package model

import (
	"time"

	"gorm.io/gorm"
)

// JenkinsConfig holds Jenkins server connection info
type JenkinsConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	URL       string         `gorm:"size:512;not null" json:"url"`
	Username  string         `gorm:"size:128;not null" json:"username"`
	APIToken  string         `gorm:"size:256;not null" json:"-"`
	Status    string         `gorm:"size:20;default:'unknown'" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JenkinsConfig) TableName() string { return "cicd_jenkins_configs" }

// JobInfo represents a Jenkins job
type JobInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
	Color       string `json:"color"`
	Buildable   bool   `json:"buildable"`
	Description string `json:"description"`
}

// BuildInfo represents a Jenkins build
type BuildInfo struct {
	Number    int   `json:"number"`
	URL       string `json:"url"`
	Result    string `json:"result"`
	Duration  int64  `json:"duration"`
	Timestamp int64  `json:"timestamp"`
	Building  bool   `json:"building"`
}

// BuildLogEntry represents log output
type BuildLogEntry struct {
	Offset  int    `json:"offset"`
	Text    string `json:"text"`
	HasMore bool   `json:"hasMore"`
}

// Pipeline is a DB-backed model representing a CI/CD pipeline
type Pipeline struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:128;not null" json:"name"`
	JenkinsConfigID uint           `gorm:"index" json:"jenkinsConfigId"`
	JobName         string         `gorm:"size:256;not null" json:"jobName"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Pipeline) TableName() string { return "cicd_pipelines" }

// PipelineRun holds a pipeline execution record
type PipelineRun struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PipelineID  uint           `gorm:"index" json:"pipelineId"`
	BuildNumber int            `json:"buildNumber"`
	Status      string         `gorm:"size:32" json:"status"`
	Log         string         `gorm:"type:longtext" json:"log,omitempty"`
	StartedAt   *time.Time     `json:"startedAt"`
	FinishedAt  *time.Time     `json:"finishedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PipelineRun) TableName() string { return "cicd_pipeline_runs" }
