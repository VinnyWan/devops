package repository

import (
	"time"

	"devops-platform/internal/modules/task/model"

	"gorm.io/gorm"
)

type ScheduleRepo struct{ db *gorm.DB }

func NewScheduleRepo(db *gorm.DB) *ScheduleRepo { return &ScheduleRepo{db: db} }

func (r *ScheduleRepo) Create(s *model.TaskSchedule) error { return r.db.Create(s).Error }

func (r *ScheduleRepo) Update(s *model.TaskSchedule) error { return r.db.Save(s).Error }

func (r *ScheduleRepo) Delete(taskID, tenantID uint) error {
	return r.db.Where("task_id = ? AND tenant_id = ?", taskID, tenantID).Delete(&model.TaskSchedule{}).Error
}

func (r *ScheduleRepo) GetByTaskID(taskID, tenantID uint) (*model.TaskSchedule, error) {
	var s model.TaskSchedule
	err := r.db.Where("task_id = ? AND tenant_id = ?", taskID, tenantID).First(&s).Error
	return &s, err
}

func (r *ScheduleRepo) ListEnabled() ([]model.TaskSchedule, error) {
	var schedules []model.TaskSchedule
	err := r.db.Where("enabled = ?", true).Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepo) SetNextRun(id uint, next time.Time) error {
	return r.db.Model(&model.TaskSchedule{}).Where("id = ?", id).Update("next_run", next).Error
}

func (r *ScheduleRepo) SetLastRun(id uint, last time.Time) error {
	return r.db.Model(&model.TaskSchedule{}).Where("id = ?", id).Update("last_run", last).Error
}
