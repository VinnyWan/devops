package model

import "time"

type Pipeline struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Branch    string    `json:"branch"`
	LastRunAt time.Time `json:"lastRunAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PipelineLog struct {
	ID         uint      `json:"id"`
	PipelineID uint      `json:"pipelineId"`
	Stage      string    `json:"stage"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
}

type TemplateStage struct {
	Name       string            `json:"name"`
	Kind       string            `json:"kind"`
	Order      int               `json:"order"`
	Parameters map[string]string `json:"parameters"`
}

type PipelineTemplate struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Source      string          `json:"source"`
	Stages      []TemplateStage `json:"stages"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type PipelineRun struct {
	ID          uint            `json:"id"`
	PipelineID  uint            `json:"pipelineId"`
	Pipeline    string          `json:"pipeline"`
	TemplateID  uint            `json:"templateId"`
	Template    string          `json:"template"`
	Branch      string          `json:"branch"`
	Environment string          `json:"environment"`
	TriggerType string          `json:"triggerType"`
	CommitID    string          `json:"commitId"`
	Operator    string          `json:"operator"`
	Status      string          `json:"status"`
	Stages      []TemplateStage `json:"stages"`
	CreatedAt   time.Time       `json:"createdAt"`
}

type JenkinsConfig struct {
	Endpoint              string    `json:"endpoint"`
	Username              string    `json:"username"`
	APIToken              string    `json:"apiToken"`
	DefaultJob            string    `json:"defaultJob"`
	TimeoutSeconds        int       `json:"timeoutSeconds"`
	TLSInsecureSkipVerify bool      `json:"tlsInsecureSkipVerify"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
