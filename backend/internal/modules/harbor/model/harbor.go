package model

import "time"

type HarborProject struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Public    bool      `json:"public"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type RepositoryImage struct {
	ID          uint      `json:"id"`
	ProjectName string    `json:"projectName"`
	Repository  string    `json:"repository"`
	Tag         string    `json:"tag"`
	Digest      string    `json:"digest"`
	Size        int64     `json:"size"`
	PushedAt    time.Time `json:"pushedAt"`
}

type HarborConfig struct {
	Endpoint              string    `json:"endpoint"`
	Project               string    `json:"project"`
	Username              string    `json:"username"`
	Password              string    `json:"password"`
	RobotToken            string    `json:"robotToken"`
	TimeoutSeconds        int       `json:"timeoutSeconds"`
	TLSInsecureSkipVerify bool      `json:"tlsInsecureSkipVerify"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
