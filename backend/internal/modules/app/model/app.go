package model

import "time"

type Application struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppTemplate struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Environment []string          `json:"environment"`
	Variables   map[string]string `json:"variables"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type ApplicationDeployment struct {
	ID           uint              `json:"id"`
	AppID        uint              `json:"appId"`
	AppName      string            `json:"appName"`
	TemplateID   uint              `json:"templateId"`
	TemplateName string            `json:"templateName"`
	Cluster      string            `json:"cluster"`
	Environment  string            `json:"environment"`
	Namespace    string            `json:"namespace"`
	Version      string            `json:"version"`
	Status       string            `json:"status"`
	Operator     string            `json:"operator"`
	Variables    map[string]string `json:"variables"`
	CreatedAt    time.Time         `json:"createdAt"`
}

type ApplicationVersion struct {
	ID          uint      `json:"id"`
	AppID       uint      `json:"appId"`
	Version     string    `json:"version"`
	Cluster     string    `json:"cluster"`
	Environment string    `json:"environment"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	Operator    string    `json:"operator"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TopologyNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Cluster  string `json:"cluster"`
	Metadata string `json:"metadata"`
}

type TopologyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type ApplicationTopology struct {
	AppID        uint           `json:"appId"`
	AppName      string         `json:"appName"`
	Environment  string         `json:"environment"`
	Nodes        []TopologyNode `json:"nodes"`
	Edges        []TopologyEdge `json:"edges"`
	LastSyncTime time.Time      `json:"lastSyncTime"`
}
