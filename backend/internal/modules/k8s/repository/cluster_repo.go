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

// Create 创建集群
func (r *ClusterRepo) Create(cluster *model.Cluster) error {
	return r.db.Create(cluster).Error
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

func (r *ClusterRepo) GetByExactName(name string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("name = ?", name).First(&cluster).Error
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

// GetByEnv 根据Env获取集群
func (r *ClusterRepo) GetByEnv(env string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("env = ?", env).First(&cluster).Error
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

// Update 更新集群
func (r *ClusterRepo) Update(cluster *model.Cluster) error {
	return r.db.Save(cluster).Error
}

// UpdateStatus 更新集群状态
func (r *ClusterRepo) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Cluster{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除集群
func (r *ClusterRepo) Delete(id uint) error {
	return r.db.Delete(&model.Cluster{}, id).Error
}

// Count 统计集群数量
func (r *ClusterRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Cluster{}).Count(&count).Error
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
