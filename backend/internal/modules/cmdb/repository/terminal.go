package repository

import (
	"strings"

	"devops-platform/internal/modules/cmdb/model"
	queryutil "devops-platform/internal/pkg/query"

	"gorm.io/gorm"
)

type TerminalRepo struct {
	db *gorm.DB
}

func NewTerminalRepo(db *gorm.DB) *TerminalRepo {
	return &TerminalRepo{db: db}
}

func (r *TerminalRepo) scopeInTenant(query *gorm.DB, tenantID uint) *gorm.DB {
	return query.Where("tenant_id = ?", tenantID)
}

func (r *TerminalRepo) Create(session *model.TerminalSession) error {
	return r.db.Create(session).Error
}

func (r *TerminalRepo) GetByIDInTenant(tenantID, id uint) (*model.TerminalSession, error) {
	var session model.TerminalSession
	if err := r.scopeInTenant(r.db, tenantID).Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *TerminalRepo) UpdateInTenant(tenantID uint, session *model.TerminalSession) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.TerminalSession
		if err := r.scopeInTenant(tx, tenantID).Where("id = ?", session.ID).First(&existing).Error; err != nil {
			return err
		}
		session.TenantID = existing.TenantID
		return tx.Save(session).Error
	})
}

func (r *TerminalRepo) ListInTenant(tenantID uint, page, pageSize int, keyword, username, status string) ([]model.TerminalSession, int64, error) {
	var sessions []model.TerminalSession
	var total int64

	query := r.scopeInTenant(r.db.Model(&model.TerminalSession{}), tenantID)
	query = queryutil.ApplyKeywordLike(query, keyword, "host_name", "host_ip")
	query = queryutil.ApplyKeywordLike(query, username, "username")

	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}
