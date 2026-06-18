package model

import (
	"time"

	"gorm.io/gorm"
)

// PrometheusConfig holds connection info for a Prometheus data source
type PrometheusConfig struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:128;not null" json:"name"`
	Endpoint       string         `gorm:"size:512;not null" json:"endpoint"`
	Username       string         `gorm:"size:128" json:"username,omitempty"`
	Password       string         `gorm:"size:256" json:"-"`
	TimeoutSeconds int            `gorm:"default:15" json:"timeoutSeconds"`
	Status         string         `gorm:"size:20;default:'unknown'" json:"status"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PrometheusConfig) TableName() string { return "monitor_prometheus_configs" }

// MetricQueryRequest is the request for querying metrics
type MetricQueryRequest struct {
	Query     string `json:"query" binding:"required"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Step      string `json:"step"`
}

// MetricResult holds a single metric data point
type MetricResult struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// MetricSeries holds a labeled time series result
type MetricSeries struct {
	Metric map[string]string `json:"metric"`
	Values []MetricResult    `json:"values"`
}

// MetricQueryResponse is the response for a metric query
type MetricQueryResponse struct {
	ResultType string         `json:"resultType"`
	Results    []MetricSeries `json:"results"`
}

// HostMetricRequest is for querying specific host metrics
type HostMetricRequest struct {
	HostID    uint   `json:"hostId" binding:"required"`
	Metric    string `json:"metric" binding:"required"` // cpu, memory, disk
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Step      string `json:"step"`
}
