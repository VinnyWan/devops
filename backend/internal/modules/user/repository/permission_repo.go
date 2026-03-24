package repository

import (
	"devops-platform/internal/modules/user/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type PermissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{
		db: db,
	}
}

// Create 创建权限
func (r *PermissionRepo) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// GetByID 根据ID获取权限
func (r *PermissionRepo) GetByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetByResourceAction 根据资源和操作获取权限
func (r *PermissionRepo) GetByResourceAction(resource, action string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// List 获取权限列表
func (r *PermissionRepo) List(page, pageSize int, resource, keyword string) ([]model.Permission, int64, error) {
	var permissions []model.Permission
	var total int64

	query := r.db.Model(&model.Permission{})

	// 按资源过滤
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	query = queryutil.ApplyKeywordLike(query, keyword, "name", "resource", "action", "description")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("resource, action").Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// ListAll 获取所有权限（不分页）
func (r *PermissionRepo) ListAll() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Order("resource, action").Find(&permissions).Error
	return permissions, err
}

// Update 更新权限
func (r *PermissionRepo) Update(permission *model.Permission) error {
	return r.db.Save(permission).Error
}

// Delete 删除权限
func (r *PermissionRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		permission := &model.Permission{ID: id}
		if err := tx.Model(permission).Association("Roles").Clear(); err != nil {
			return err
		}
		return tx.Delete(permission).Error
	})
}

// GetPermissionsByIDs 批量获取权限
func (r *PermissionRepo) GetPermissionsByIDs(ids []uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Find(&permissions, ids).Error
	return permissions, err
}

// GetPermissionsByResource 获取指定资源的所有权限
func (r *PermissionRepo) GetPermissionsByResource(resource string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Where("resource = ?", resource).Order("action").Find(&permissions).Error
	return permissions, err
}
