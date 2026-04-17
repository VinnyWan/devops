package repository

import (
	"devops-platform/internal/modules/cmdb/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type CredentialRepo struct {
	db *gorm.DB
}

func NewCredentialRepo(db *gorm.DB) *CredentialRepo {
	return &CredentialRepo{db: db}
}

func (r *CredentialRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	if tenantID == 0 {
		return query
	}
	return query.Where("tenant_id = ?", tenantID)
}

func (r *CredentialRepo) Create(cred *model.Credential) error {
	return r.db.Create(cred).Error
}

func (r *CredentialRepo) CreateInTenant(tenantID uint, cred *model.Credential) error {
	if tenantID > 0 {
		cred.TenantID = &tenantID
	}
	return r.Create(cred)
}

func (r *CredentialRepo) GetByID(id uint) (*model.Credential, error) {
	var cred model.Credential
	if err := r.db.First(&cred, id).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepo) GetByIDInTenant(tenantID uint, id uint) (*model.Credential, error) {
	if tenantID == 0 {
		return r.GetByID(id)
	}
	var cred model.Credential
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&cred).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

// ListInTenant 返回不含敏感字段的凭据列表
func (r *CredentialRepo) ListInTenant(tenantID uint, page, pageSize int, keyword string) ([]model.Credential, int64, error) {
	var creds []model.Credential
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.Credential{}), tenantID)
	query = queryutil.ApplyKeywordLike(query, keyword, "name", "username", "description")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Select("id, tenant_id, name, type, username, description, created_at, updated_at").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&creds).Error; err != nil {
		return nil, 0, err
	}

	return creds, total, nil
}

func (r *CredentialRepo) Update(cred *model.Credential) error {
	return r.db.Save(cred).Error
}

func (r *CredentialRepo) UpdateInTenant(tenantID uint, cred *model.Credential) error {
	if tenantID == 0 {
		return r.Update(cred)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.Credential
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, cred.ID).First(&existing).Error; err != nil {
			return err
		}
		return tx.Save(cred).Error
	})
}

func (r *CredentialRepo) DeleteInTenant(tenantID uint, id uint) error {
	if tenantID == 0 {
		return r.db.Delete(&model.Credential{}, id).Error
	}
	return r.scopeInTenant(r.db, tenantID).Where("id = ?", id).Delete(&model.Credential{}).Error
}

// IsReferencedByHosts 检查凭据是否被主机引用
func (r *CredentialRepo) IsReferencedByHosts(credentialID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Host{}).Where("credential_id = ?", credentialID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
