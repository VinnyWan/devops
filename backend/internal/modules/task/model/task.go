package model

import (
	"time"

	"gorm.io/gorm"
)

// TaskType defines the execution type of a task.
type TaskType string

const (
	TaskTypeShell   TaskType = "shell"
	TaskTypePython  TaskType = "python"
	TaskTypeAnsible TaskType = "ansible"
)

// TaskStatus represents the current state of a task execution.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusTimeout    TaskStatus = "timeout"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task is a reusable task definition.
type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenant_id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"size:1024" json:"description"`
	Type        TaskType       `gorm:"size:32;not null" json:"type"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Timeout     int            `gorm:"default:300" json:"timeout"`
	CreatedBy   uint           `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Task) TableName() string { return "tasks" }

// TaskExecution records a single run of a task.
type TaskExecution struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	TaskID     uint       `gorm:"index;not null" json:"task_id"`
	TenantID   uint       `gorm:"index;not null" json:"tenant_id"`
	Status     TaskStatus `gorm:"size:32;not null;default:pending" json:"status"`
	Targets    string     `gorm:"type:text" json:"targets"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	DurationMs int64      `json:"duration_ms"`
	Result     string     `gorm:"type:text" json:"result"`
	LogPath    string     `gorm:"size:512" json:"log_path"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (TaskExecution) TableName() string { return "task_executions" }

// TaskSchedule defines a cron-based schedule for recurring task execution.
type TaskSchedule struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TaskID    uint           `gorm:"uniqueIndex;not null" json:"task_id"`
	TenantID  uint          `gorm:"index;not null" json:"tenant_id"`
	CronExpr  string         `gorm:"size:128;not null" json:"cron_expr"`
	Enabled   bool           `gorm:"default:true" json:"enabled"`
	NextRun   *time.Time     `json:"next_run"`
	LastRun   *time.Time     `json:"last_run"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TaskSchedule) TableName() string { return "task_schedules" }
