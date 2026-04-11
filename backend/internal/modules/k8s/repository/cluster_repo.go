package repository

import (
	"devops-platform/internal/modules/k8s/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type ClusterRepo struct {
	db *gorm.DB
}

func NewClusterRepo(db *gorm.DB) *ClusterRepo {
	return &ClusterRepo{
		db: db,
	}
}

func (r *ClusterRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	if tenantID == 0 {
		return query
	}
	return query.Where("tenant_id = ?", tenantID)
}

// Create 创建集群
func (r *ClusterRepo) Create(cluster *model.Cluster) error {
	return r.db.Create(cluster).Error
}

func (r *ClusterRepo) CreateInTenant(tenantID uint, cluster *model.Cluster) error {
	if tenantID > 0 {
		cluster.TenantID = &tenantID
	}
	return r.Create(cluster)
}

// GetByID 根据ID获取集群
func (r *ClusterRepo) GetByID(id uint) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.First(&cluster, id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetByIDInTenant(tenantID uint, id uint) (*model.Cluster, error) {
	if tenantID == 0 {
		return r.GetByID(id)
	}
	var cluster model.Cluster
	err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetByExactName(name string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetByExactNameInTenant(tenantID uint, name string) (*model.Cluster, error) {
	if tenantID == 0 {
		return r.GetByExactName(name)
	}
	var cluster model.Cluster
	err := r.scopeInTenant(r.db, tenantID).Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// GetByName 根据名称获取集群
func (r *ClusterRepo) GetByName(name string) ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := r.db.
		Where("name LIKE ?", "%"+name+"%").
		Find(&clusters).Error
	return clusters, err
}

func (r *ClusterRepo) GetByNameInTenant(tenantID uint, name string) ([]model.Cluster, error) {
	if tenantID == 0 {
		return r.GetByName(name)
	}
	var clusters []model.Cluster
	err := r.scopeInTenant(r.db, tenantID).
		Where("name LIKE ?", "%"+name+"%").
		Find(&clusters).Error
	return clusters, err
}

// GetByEnv 根据Env获取集群
func (r *ClusterRepo) GetByEnv(env string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("env = ?", env).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetByEnvInTenant(tenantID uint, env string) (*model.Cluster, error) {
	if tenantID == 0 {
		return r.GetByEnv(env)
	}
	var cluster model.Cluster
	err := r.scopeInTenant(r.db, tenantID).Where("env = ?", env).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetDefault() (*model.Cluster, error) {
	var cluster model.Cluster
	if err := r.db.Where("is_default = ?", true).Order("id DESC").First(&cluster).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) GetDefaultInTenant(tenantID uint) (*model.Cluster, error) {
	if tenantID == 0 {
		return r.GetDefault()
	}
	var cluster model.Cluster
	if err := r.scopeInTenant(r.db, tenantID).
		Where("is_default = ?", true).
		Order("id DESC").
		First(&cluster).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepo) SetDefault(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Cluster{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Cluster{}).Where("id = ?", id).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ClusterRepo) SetDefaultInTenant(tenantID, id uint) error {
	if tenantID == 0 {
		return r.SetDefault(id)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Cluster{}).
			Where("tenant_id = ? AND is_default = ?", tenantID, true).
			Update("is_default", false).Error; err != nil {
			return err
		}
		result := tx.Model(&model.Cluster{}).
			Where("tenant_id = ? AND id = ?", tenantID, id).
			Update("is_default", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// List 获取集群列表
func (r *ClusterRepo) List(page, pageSize int, env, keyword string) ([]model.Cluster, int64, error) {
	var clusters []model.Cluster
	var total int64

	query := r.db.Model(&model.Cluster{})

	// 环境过滤
	if env != "" {
		query = query.Where("env = ?", env)
	}

	query = queryutil.ApplyKeywordLike(query, keyword, "name", "url", "remark", "labels", "status", "env", "k8s_version")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&clusters).Error; err != nil {
		return nil, 0, err
	}

	return clusters, total, nil
}

func (r *ClusterRepo) ListInTenant(tenantID uint, page, pageSize int, env, keyword string) ([]model.Cluster, int64, error) {
	if tenantID == 0 {
		return r.List(page, pageSize, env, keyword)
	}
	var clusters []model.Cluster
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.Cluster{}), tenantID)
	if env != "" {
		query = query.Where("env = ?", env)
	}
	query = queryutil.ApplyKeywordLike(query, keyword, "name", "url", "remark", "labels", "status", "env", "k8s_version")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&clusters).Error; err != nil {
		return nil, 0, err
	}
	return clusters, total, nil
}

// Update 更新集群
func (r *ClusterRepo) Update(cluster *model.Cluster) error {
	return r.db.Save(cluster).Error
}

func (r *ClusterRepo) UpdateInTenant(tenantID uint, cluster *model.Cluster) error {
	if tenantID == 0 {
		return r.Update(cluster)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.Cluster
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, cluster.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(cluster).Error
	})
}

// UpdateStatus 更新集群状态
func (r *ClusterRepo) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Cluster{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ClusterRepo) UpdateStatusInTenant(tenantID uint, id uint, status string) error {
	if tenantID == 0 {
		return r.UpdateStatus(id, status)
	}
	return r.db.Model(&model.Cluster{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("status", status).Error
}

// Delete 删除集群
func (r *ClusterRepo) Delete(id uint) error {
	return r.db.Delete(&model.Cluster{}, id).Error
}

func (r *ClusterRepo) DeleteInTenant(tenantID uint, id uint) error {
	if tenantID == 0 {
		return r.Delete(id)
	}
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.Cluster{}).Error
}

// Count 统计集群数量
func (r *ClusterRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Cluster{}).Count(&count).Error
	return count, err
}

func (r *ClusterRepo) CountInTenant(tenantID uint) (int64, error) {
	if tenantID == 0 {
		return r.Count()
	}
	var count int64
	err := r.scopeInTenant(r.db.Model(&model.Cluster{}), tenantID).Count(&count).Error
	return count, err
}

func (r *ClusterRepo) Search(
	name string,
	env string,
	page int,
	pageSize int,
) ([]model.Cluster, int64, error) {

	var (
		clusters []model.Cluster
		total    int64
	)

	db := r.db.Model(&model.Cluster{})

	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	if env != "" {
		db = db.Where("env = ?", env)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := db.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("id DESC").
		Find(&clusters).Error

	return clusters, total, err
}

func (r *ClusterRepo) SearchInTenant(
	tenantID uint,
	name string,
	env string,
	page int,
	pageSize int,
) ([]model.Cluster, int64, error) {
	if tenantID == 0 {
		return r.Search(name, env, page, pageSize)
	}
	var (
		clusters []model.Cluster
		total    int64
	)

	db := r.scopeInTenant(r.db.Model(&model.Cluster{}), tenantID)
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if env != "" {
		db = db.Where("env = ?", env)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("id DESC").
		Find(&clusters).Error
	return clusters, total, err
}
