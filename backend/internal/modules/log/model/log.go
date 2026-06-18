package model

import (
	"time"

	"gorm.io/gorm"
)

// LogSource holds connection info for a log backend
type LogSource struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	Type         string         `gorm:"size:32;not null;default:'elasticsearch'" json:"type"`
	Endpoint     string         `gorm:"size:512;not null" json:"endpoint"`
	Username     string         `gorm:"size:128" json:"username,omitempty"`
	Password     string         `gorm:"size:256" json:"-"`
	IndexPattern string         `gorm:"size:256;not null;default:'app-logs-*'" json:"indexPattern"`
	Status       string         `gorm:"size:20;default:'unknown'" json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LogSource) TableName() string { return "log_sources" }

// SearchRequest for log search
type SearchRequest struct {
	SourceID  uint     `json:"sourceId"`
	Keywords  []string `json:"keywords"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
	Level     string   `json:"level"`   // ERROR, WARN, INFO, DEBUG
	Service   string   `json:"service"` // source service name
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
	SortOrder string   `json:"sortOrder"` // asc, desc
}

// LogEntry represents a single log record
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Host      string `json:"host,omitempty"`
	TraceID   string `json:"traceId,omitempty"`
}

// SearchResponse for log search results
type SearchResponse struct {
	Entries    []LogEntry `json:"entries"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}
