package model

import (
	"time"

	"gorm.io/gorm"
)

// DbConnection stores a managed database connection configuration.
type DbConnection struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenantId"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Type        string         `gorm:"size:20;not null;comment:mysql|postgresql" json:"type"`
	Host        string         `gorm:"size:255;not null" json:"host"`
	Port        int            `gorm:"default:3306" json:"port"`
	Database    string         `gorm:"size:128" json:"database"`
	Username    string         `gorm:"size:128;not null" json:"username"`
	Password    string         `gorm:"size:500;not null" json:"-"`
	Mode        string         `gorm:"size:16;default:read_write;comment:read_only|read_write" json:"mode"`
	Status      string         `gorm:"size:20;default:active" json:"status"`
	Description string         `gorm:"size:512" json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DbConnection) TableName() string { return "db_connections" }

// SqlRecord stores an audit log entry for executed SQL.
type SqlRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     uint      `gorm:"index;not null" json:"tenantId"`
	ConnectionID uint      `gorm:"index;not null" json:"connectionId"`
	UserID       uint      `json:"userId"`
	Database     string    `gorm:"size:128" json:"database"`
	SQL          string    `gorm:"type:text;not null" json:"sql"`
	Mode         string    `gorm:"size:16" json:"mode"`
	Sensitive    bool      `json:"sensitive"`
	RiskLevel    string    `gorm:"size:16" json:"riskLevel"`
	Duration     int64     `json:"duration"`
	RowsAffected int64     `json:"rowsAffected"`
	Error        string    `gorm:"type:text" json:"error"`
	ClientIP     string    `gorm:"size:45" json:"clientIp"`
	ExecutedAt   time.Time `json:"executedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (SqlRecord) TableName() string { return "sql_records" }

// Risk levels for sensitive SQL detection.
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

// SensitivePatterns defines SQL patterns that require extra scrutiny.
var SensitivePatterns = []struct {
	Pattern  string
	Risk     string
	Label    string
}{
	{Pattern: `(?i)\bDROP\s+(TABLE|DATABASE|INDEX)`, Risk: RiskHigh, Label: "DROP 操作"},
	{Pattern: `(?i)\bTRUNCATE\s+`, Risk: RiskHigh, Label: "TRUNCATE 操作"},
	{Pattern: `(?i)\bDELETE\s+FROM\s+\S+(?!\s+WHERE)`, Risk: RiskHigh, Label: "DELETE 无 WHERE"},
	{Pattern: `(?i)\bDELETE\s+FROM\s+\S+\s+WHERE\b`, Risk: RiskMedium, Label: "DELETE 有 WHERE"},
	{Pattern: `(?i)\bALTER\s+TABLE`, Risk: RiskMedium, Label: "ALTER TABLE"},
	{Pattern: `(?i)\bUPDATE\s+\S+\s+SET\b(?!.*\bWHERE\b)`, Risk: RiskHigh, Label: "UPDATE 无 WHERE"},
	{Pattern: `(?i)\bUPDATE\s+\S+\s+SET\b.*\bWHERE\b`, Risk: RiskLow, Label: "UPDATE 有 WHERE"},
	{Pattern: `(?i)\bGRANT\s+`, Risk: RiskHigh, Label: "GRANT 操作"},
	{Pattern: `(?i)\bREVOKE\s+`, Risk: RiskHigh, Label: "REVOKE 操作"},
}
