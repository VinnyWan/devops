package repository

import (
	"devops-platform/internal/modules/cmdb/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type HostRepo struct {
	db *gorm.DB
}

func NewHostRepo(db *gorm.DB) *HostRepo {
	return &HostRepo{db: db}
}

func (r *HostRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	if tenantID == 0 {
		return query
	}
	return query.Where("tenant_id = ?", tenantID)
}

func (r *HostRepo) Create(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *HostRepo) CreateInTenant(tenantID uint, host *model.Host) error {
	if tenantID > 0 {
		host.TenantID = &tenantID
	}
	return r.Create(host)
}

func (r *HostRepo) GetByID(id uint) (*model.Host, error) {
	var host model.Host
	if err := r.db.First(&host, id).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepo) GetByIDInTenant(tenantID uint, id uint) (*model.Host, error) {
	if tenantID == 0 {
		return r.GetByID(id)
	}
	var host model.Host
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *HostRepo) ListInTenant(tenantID uint, page, pageSize int, groupID uint, status, keyword string, allowedHostIDs []uint) ([]model.Host, int64, error) {
	var hosts []model.Host
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID)

	if len(allowedHostIDs) > 0 {
		query = query.Where("id IN ?", allowedHostIDs)
	}

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = queryutil.ApplyKeywordLike(query, keyword, "hostname", "ip", "os_name", "description")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&hosts).Error; err != nil {
		return nil, 0, err
	}

	return hosts, total, nil
}

func (r *HostRepo) Update(host *model.Host) error {
	return r.db.Save(host).Error
}

func (r *HostRepo) UpdateInTenant(tenantID uint, host *model.Host) error {
	if tenantID == 0 {
		return r.Update(host)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.Host
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, host.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(host).Error
	})
}

func (r *HostRepo) DeleteInTenant(tenantID uint, id uint) error {
	if tenantID == 0 {
		return r.db.Delete(&model.Host{}, id).Error
	}
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.Host{}).Error
}

func (r *HostRepo) CountByStatusInTenant(tenantID uint) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID)
	if err := query.Select("status, count(*) as count").Group("status").Find(&results).Error; err != nil {
		return nil, err
	}

	statusMap := make(map[string]int64)
	for _, r := range results {
		statusMap[r.Status] = r.Count
	}
	return statusMap, nil
}

func (r *HostRepo) CountByGroupInTenant(tenantID uint) (map[uint]int64, error) {
	type result struct {
		GroupID uint
		Count   int64
	}
	var results []result
	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID)
	if err := query.Select("group_id, count(*) as count").Group("group_id").Find(&results).Error; err != nil {
		return nil, err
	}

	groupMap := make(map[uint]int64)
	for _, r := range results {
		groupMap[r.GroupID] = r.Count
	}
	return groupMap, nil
}

func (r *HostRepo) CountInTenant(tenantID uint) (int64, error) {
	var count int64
	query := r.scopeInTenant(r.db.Model(&model.Host{}), tenantID)
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *HostRepo) BatchCreateInTenant(tenantID uint, hosts []model.Host) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range hosts {
			if tenantID > 0 {
				hosts[i].TenantID = &tenantID
			}
			if err := tx.Create(&hosts[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
