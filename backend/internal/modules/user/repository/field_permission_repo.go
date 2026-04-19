package repository

import (
	"devops-platform/internal/modules/user/model"

	"gorm.io/gorm"
)

type FieldPermissionRepo struct {
	db *gorm.DB
}

func NewFieldPermissionRepo(db *gorm.DB) *FieldPermissionRepo {
	return &FieldPermissionRepo{db: db}
}

// GetByRoleIDs 批量查询多个角色的字段权限
func (r *FieldPermissionRepo) GetByRoleIDs(roleIDs []uint) ([]model.FieldPermission, error) {
	var fps []model.FieldPermission
	err := r.db.Where("role_id IN ?", roleIDs).Find(&fps).Error
	return fps, err
}

// Upsert 创建或更新字段权限
func (r *FieldPermissionRepo) Upsert(fp *model.FieldPermission) error {
	return r.db.Save(fp).Error
}

// DeleteByRoleID 删除角色的所有字段权限
func (r *FieldPermissionRepo) DeleteByRoleID(roleID uint) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.FieldPermission{}).Error
}

// GetByRoleIDAndResource 查询角色对某资源的字段权限
func (r *FieldPermissionRepo) GetByRoleIDAndResource(roleID uint, resource string) ([]model.FieldPermission, error) {
	var fps []model.FieldPermission
	err := r.db.Where("role_id = ? AND resource = ?", roleID, resource).Find(&fps).Error
	return fps, err
}
