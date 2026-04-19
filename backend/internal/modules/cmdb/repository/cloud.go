package repository

import (
	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CloudRepo struct {
	db *gorm.DB
}

func NewCloudRepo(db *gorm.DB) *CloudRepo {
	return &CloudRepo{db: db}
}

func (r *CloudRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

// CloudAccount CRUD

func (r *CloudRepo) CreateAccount(account *model.CloudAccount) error {
	return r.db.Create(account).Error
}

func (r *CloudRepo) GetAccountByIDInTenant(tenantID, id uint) (*model.CloudAccount, error) {
	var account model.CloudAccount
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *CloudRepo) UpdateAccount(account *model.CloudAccount) error {
	return r.db.Save(account).Error
}

func (r *CloudRepo) DeleteAccountInTenant(tenantID, id uint) error {
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.CloudAccount{}).Error
}

func (r *CloudRepo) ListAccountsInTenant(tenantID uint, page, pageSize int, status string) ([]model.CloudAccount, int64, error) {
	var accounts []model.CloudAccount
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.CloudAccount{}), tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *CloudRepo) ListAllActiveAccounts() ([]model.CloudAccount, error) {
	var accounts []model.CloudAccount
	if err := r.db.Where("status = ?", "active").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// CloudResource CRUD

func (r *CloudRepo) UpsertResource(resource *model.CloudResource) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "cloud_account_id"},
			{Name: "resource_type"},
			{Name: "resource_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "region", "zone", "state", "spec", "synced_at", "updated_at",
		}),
	}).Create(resource).Error
}

func (r *CloudRepo) ListResourcesByAccountInTenant(tenantID, accountID uint, resourceType string, page, pageSize int) ([]model.CloudResource, int64, error) {
	var resources []model.CloudResource
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.CloudResource{}), tenantID).
		Where("cloud_account_id = ?", accountID)
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("synced_at DESC").Offset(offset).Limit(pageSize).Find(&resources).Error; err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *CloudRepo) DeleteResourcesByAccount(accountID uint) error {
	return r.db.Where("cloud_account_id = ?", accountID).Delete(&model.CloudResource{}).Error
}

func (r *CloudRepo) GetHostByCloudInstanceID(tenantID uint, instanceID string) (*model.Host, error) {
	var host model.Host
	query := r.scopeInTenant(r.db, tenantID).Where("cloud_instance_id = ?", instanceID)
	if err := query.First(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *CloudRepo) CreateHost(host *model.Host) error {
	return r.db.Create(host).Error
}

func (r *CloudRepo) UpdateHost(host *model.Host) error {
	return r.db.Save(host).Error
}
