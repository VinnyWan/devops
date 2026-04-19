package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type PermissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

func (r *PermissionRepo) Create(perm *model.HostPermission) error {
	return r.db.Create(perm).Error
}

func (r *PermissionRepo) GetByIDInTenant(tenantID, id uint) (*model.HostPermission, error) {
	var perm model.HostPermission
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *PermissionRepo) UpdateInTenant(tenantID uint, perm *model.HostPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.HostPermission
		if err := r.scopeInTenant(tx, tenantID).Where("id = ?", perm.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(perm).Error
	})
}

func (r *PermissionRepo) DeleteInTenant(tenantID, id uint) error {
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.HostPermission{}).Error
}

func (r *PermissionRepo) ListInTenant(tenantID uint, page, pageSize int, userID, hostGroupID uint, permission string) ([]model.HostPermission, int64, error) {
	var perms []model.HostPermission
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.HostPermission{}), tenantID)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if hostGroupID > 0 {
		query = query.Where("host_group_id = ?", hostGroupID)
	}
	if permission != "" {
		query = query.Where("permission = ?", permission)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&perms).Error; err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

func (r *PermissionRepo) GetByUserInTenant(tenantID, userID uint) ([]model.HostPermission, error) {
	var perms []model.HostPermission
	if err := r.scopeInTenant(r.db, tenantID).Where("user_id = ?", userID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *PermissionRepo) ExistsByUserGroupPermission(tenantID, userID, hostGroupID uint, permission string) (bool, error) {
	var count int64
	if err := r.scopeInTenant(r.db.Model(&model.HostPermission{}), tenantID).
		Where("user_id = ? AND host_group_id = ? AND permission = ?", userID, hostGroupID, permission).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
