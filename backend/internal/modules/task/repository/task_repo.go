package repository

import (
	"devops-platform/internal/modules/task/model"

	"gorm.io/gorm"
)

type TaskRepo struct{ db *gorm.DB }

func NewTaskRepo(db *gorm.DB) *TaskRepo { return &TaskRepo{db: db} }

func (r *TaskRepo) Create(t *model.Task) error { return r.db.Create(t).Error }

func (r *TaskRepo) Update(t *model.Task) error { return r.db.Save(t).Error }

func (r *TaskRepo) Delete(id, tenantID uint) error {
	return r.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Task{}).Error
}

func (r *TaskRepo) GetByID(id, tenantID uint) (*model.Task, error) {
	var t model.Task
	err := r.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&t).Error
	return &t, err
}

func (r *TaskRepo) List(tenantID uint, keyword string, page, pageSize int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64
	q := r.db.Model(&model.Task{}).Where("tenant_id = ?", tenantID)
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}
