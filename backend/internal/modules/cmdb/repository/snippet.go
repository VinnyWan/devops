package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type SnippetRepo struct {
	db *gorm.DB
}

func NewSnippetRepo(db *gorm.DB) *SnippetRepo {
	return &SnippetRepo{db: db}
}

func (r *SnippetRepo) ListInTenant(tenantID uint, userID uint, page, pageSize int) ([]model.CommandSnippet, int64, error) {
	var list []model.CommandSnippet
	var total int64

	query := r.db.Where("tenant_id = ? AND (visibility = 'public' OR visibility = 'team' OR user_id = ?)", tenantID, userID)
	query.Model(&model.CommandSnippet{}).Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *SnippetRepo) GetByID(id uint) (*model.CommandSnippet, error) {
	var s model.CommandSnippet
	err := r.db.First(&s, id).Error
	return &s, err
}

func (r *SnippetRepo) Create(snippet *model.CommandSnippet) error {
	return r.db.Create(snippet).Error
}

func (r *SnippetRepo) Update(snippet *model.CommandSnippet) error {
	return r.db.Save(snippet).Error
}

func (r *SnippetRepo) Delete(id uint) error {
	return r.db.Delete(&model.CommandSnippet{}, id).Error
}

func (r *SnippetRepo) Search(tenantID uint, userID uint, keyword string, limit int) ([]model.CommandSnippet, error) {
	var list []model.CommandSnippet
	pattern := "%" + keyword + "%"
	err := r.db.Where(
		"tenant_id = ? AND (visibility = 'public' OR visibility = 'team' OR user_id = ?)", tenantID, userID,
	).Where(
		"name LIKE ? OR tags LIKE ? OR content LIKE ?", pattern, pattern, pattern,
	).Limit(limit).Find(&list).Error
	return list, err
}
