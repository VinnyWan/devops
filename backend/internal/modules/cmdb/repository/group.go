package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type GroupRepo struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

func (r *GroupRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	if tenantID == 0 {
		return query
	}
	return query.Where("tenant_id = ?", tenantID)
}

func (r *GroupRepo) Create(group *model.HostGroup) error {
	return r.db.Create(group).Error
}

func (r *GroupRepo) CreateInTenant(tenantID uint, group *model.HostGroup) error {
	if tenantID > 0 {
		group.TenantID = &tenantID
	}
	return r.Create(group)
}

func (r *HostRepo) GetByGroupIDInTenant(tenantID uint, groupID uint, page, pageSize int) ([]model.Host, int64, error) {
	var hosts []model.Host
	var total int64
	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID).Where("group_id = ?", groupID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&hosts).Error; err != nil {
		return nil, 0, err
	}
	return hosts, total, nil
}

func (r *GroupRepo) GetByID(id uint) (*model.HostGroup, error) {
	var group model.HostGroup
	if err := r.db.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepo) GetByIDInTenant(tenantID uint, id uint) (*model.HostGroup, error) {
	if tenantID == 0 {
		return r.GetByID(id)
	}
	var group model.HostGroup
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepo) ListInTenant(tenantID uint) ([]model.HostGroup, error) {
	var groups []model.HostGroup
	query := r.scopeInTenant(r.db.Model(&model.HostGroup{}), tenantID)
	if err := query.Order("level ASC, sort_order ASC, created_at ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepo) GetChildren(parentID uint) ([]model.HostGroup, error) {
	var groups []model.HostGroup
	if err := r.db.Where("parent_id = ?", parentID).Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepo) HasChildren(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.HostGroup{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GroupRepo) Update(group *model.HostGroup) error {
	return r.db.Save(group).Error
}

func (r *GroupRepo) UpdateInTenant(tenantID uint, group *model.HostGroup) error {
	if tenantID == 0 {
		return r.Update(group)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.HostGroup
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, group.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(group).Error
	})
}

func (r *GroupRepo) DeleteInTenant(tenantID uint, id uint) error {
	if tenantID == 0 {
		return r.db.Delete(&model.HostGroup{}, id).Error
	}
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.HostGroup{}).Error
}
