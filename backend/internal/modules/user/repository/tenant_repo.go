package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type TenantRepo struct {
	db *gorm.DB
}

func NewTenantRepo(db *gorm.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) Create(tenant *model.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *TenantRepo) GetByID(id uint) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepo) GetByCode(code string) (*model.Tenant, error) {
	var tenant model.Tenant
	if err := r.db.Where("code = ?", code).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepo) List(page, pageSize int, keyword, status string) ([]model.Tenant, int64, error) {
	var (
		items []model.Tenant
		total int64
	)

	query := r.db.Model(&model.Tenant{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = queryutil.ApplyKeywordLike(query, keyword, "name", "code", "description", "contact_name", "contact_email")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *TenantRepo) UpdateByID(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(updates).Error
}
