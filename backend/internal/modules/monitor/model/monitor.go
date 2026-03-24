package model

import "time"

type QueryRequest struct {
	Metric string
	Start  time.Time
	End    time.Time
	Step   string
}

type QueryPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type QuerySeries struct {
	Labels map[string]string `json:"labels"`
	Points []QueryPoint      `json:"points"`
}

type QueryResult struct {
	Metric string        `json:"metric"`
	Start  time.Time     `json:"start"`
	End    time.Time     `json:"end"`
	Step   string        `json:"step"`
	Series []QuerySeries `json:"series"`
}

type PrometheusConfig struct {
	Endpoint              string    `json:"endpoint"`
	QueryPath             string    `json:"queryPath"`
	TimeoutSeconds        int       `json:"timeoutSeconds"`
	Username              string    `json:"username"`
	Password              string    `json:"password"`
	BearerToken           string    `json:"bearerToken"`
	TLSInsecureSkipVerify bool      `json:"tlsInsecureSkipVerify"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
