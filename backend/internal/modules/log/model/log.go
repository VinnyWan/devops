package model

import "time"

type LogEntry struct {
	ID        uint      `json:"id"`
	Cluster   string    `json:"cluster"`
	Namespace string    `json:"namespace"`
	Pod       string    `json:"pod"`
	Source    string    `json:"source"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type SearchRequest struct {
	Keyword  string
	Source   string
	Level    string
	Start    time.Time
	End      time.Time
	Page     int
	PageSize int
}
