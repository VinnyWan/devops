package model

import (
	"time"

	"gorm.io/gorm"
)

type Cluster struct {
	ID         uint      `gorm:"primaryKey"`
	Name       string    `gorm:"uniqueIndex;size:100;not null"`
	Url        string    `gorm:"size:255;not null"`
	AuthType   string    `gorm:"size:20;not null"`
	Kubeconfig string    `gorm:"type:text"`
	Token      string    `gorm:"size:500"`
	CaData     string    `gorm:"type:text"`
	Status     string    `gorm:"size:20;default:'pending';index"`
	K8sVersion string    `gorm:"size:50"`
	NodeCount  int       `gorm:"default:0"`
	IsDefault  bool      `gorm:"default:false;index" json:"isDefault"`
	Remark     string    `gorm:"type:varchar(255);charset:utf8mb4"`
	Labels     string    `gorm:"size:500"`
	Env        string    `gorm:"size:20;index"`
	CreatedAt  time.Time `gorm:"index"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
