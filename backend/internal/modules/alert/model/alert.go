package model

import "time"

type Rule struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Expr        string    `json:"expr"`
	Severity    string    `json:"severity"`
	Enabled     bool      `json:"enabled"`
	Cluster     string    `json:"cluster"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Description string    `json:"description"`
}

type Silence struct {
	ID        uint      `json:"id"`
	RuleID    uint      `json:"ruleId"`
	Reason    string    `json:"reason"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NotificationChannel struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Target    string    `json:"target"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type History struct {
	ID          uint      `json:"id"`
	RuleID      uint      `json:"ruleId"`
	RuleName    string    `json:"ruleName"`
	Status      string    `json:"status"`
	Severity    string    `json:"severity"`
	Summary     string    `json:"summary"`
	StartsAt    time.Time `json:"startsAt"`
	EndsAt      time.Time `json:"endsAt"`
	Cluster     string    `json:"cluster"`
	Namespace   string    `json:"namespace"`
	Instance    string    `json:"instance"`
	Fingerprint string    `json:"fingerprint"`
}

type AlertmanagerConfig struct {
	Endpoint              string    `json:"endpoint"`
	APIPath               string    `json:"apiPath"`
	TimeoutSeconds        int       `json:"timeoutSeconds"`
	Username              string    `json:"username"`
	Password              string    `json:"password"`
	BearerToken           string    `json:"bearerToken"`
	TLSInsecureSkipVerify bool      `json:"tlsInsecureSkipVerify"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
