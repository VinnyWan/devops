package repository

import (
	"devops-platform/internal/modules/user/model"

	"gorm.io/gorm"
)

type ApiKeyRepo struct {
	db *gorm.DB
}

func NewApiKeyRepo(db *gorm.DB) *ApiKeyRepo {
	return &ApiKeyRepo{db: db}
}

func (r *ApiKeyRepo) Create(apiKey *model.ApiKey) error {
	return r.db.Create(apiKey).Error
}

func (r *ApiKeyRepo) List(userID, tenantID uint, keyword string, page, pageSize int) ([]model.ApiKey, int64, error) {
	var keys []model.ApiKey
	var total int64

	query := r.db.Model(&model.ApiKey{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

func (r *ApiKeyRepo) GetByID(id, userID, tenantID uint) (*model.ApiKey, error) {
	var key model.ApiKey
	err := r.db.Where("id = ? AND user_id = ? AND tenant_id = ?", id, userID, tenantID).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *ApiKeyRepo) Delete(id, userID, tenantID uint) error {
	return r.db.Where("id = ? AND user_id = ? AND tenant_id = ?", id, userID, tenantID).Delete(&model.ApiKey{}).Error
}

func (r *ApiKeyRepo) FindByHash(hash string) (*model.ApiKey, error) {
	var key model.ApiKey
	err := r.db.Where("key_hash = ?", hash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *ApiKeyRepo) UpdateLastUsed(id uint) {
	r.db.Model(&model.ApiKey{}).Where("id = ?", id).Update("last_used", gorm.Expr("NOW()"))
}
