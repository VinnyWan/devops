package repository

import (
	"time"

	"devops-platform/internal/modules/task/model"

	"gorm.io/gorm"
)

type ExecutionRepo struct{ db *gorm.DB }

func NewExecutionRepo(db *gorm.DB) *ExecutionRepo { return &ExecutionRepo{db: db} }

func (r *ExecutionRepo) Create(e *model.TaskExecution) error { return r.db.Create(e).Error }

func (r *ExecutionRepo) UpdateStatus(id uint, status model.TaskStatus, result string, durationMs int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      status,
		"finished_at": now,
		"duration_ms": durationMs,
	}
	if result != "" {
		updates["result"] = result
	}
	return r.db.Model(&model.TaskExecution{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ExecutionRepo) SetStarted(id uint) error {
	now := time.Now()
	return r.db.Model(&model.TaskExecution{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.TaskStatusRunning,
		"started_at": now,
	}).Error
}

func (r *ExecutionRepo) GetByID(id, tenantID uint) (*model.TaskExecution, error) {
	var e model.TaskExecution
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&e).Error
	return &e, err
}

func (r *ExecutionRepo) ListByTask(taskID, tenantID uint, page, pageSize int) ([]model.TaskExecution, int64, error) {
	var execs []model.TaskExecution
	var total int64
	q := r.db.Model(&model.TaskExecution{}).Where("task_id = ? AND tenant_id = ?", taskID, tenantID)
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&execs).Error
	return execs, total, err
}

func (r *ExecutionRepo) DeleteOlderThan(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return r.db.Where("created_at < ?", cutoff).Delete(&model.TaskExecution{}).Error
}
